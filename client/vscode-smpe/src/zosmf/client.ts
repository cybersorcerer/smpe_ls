/**
 * z/OSMF HTTP Client
 * Handles REST API communication with z/OSMF SMP/E endpoints
 */

import * as https from 'https';
import * as http from 'http';
import * as vscode from 'vscode';
import {
    ZosmfServer,
    Credentials,
    SysmodQueryRequest,
    DddefQueryRequest,
    AsyncResponse,
    StatusResponse,
    QueryResult,
    ProgressCallback,
    UssDirectoryListing,
    DatasetMemberListing
} from './types';

const POLL_INTERVAL_MS = 2000;
const MAX_404_RETRIES = 30;    // retry up to 1 minute while z/OSMF initializes the query

export class ZosmfClient {
    private outputChannel: vscode.OutputChannel;
    private insecureAgent: https.Agent;
    private maxPollAttempts: number;

    constructor(outputChannel: vscode.OutputChannel, queryTimeoutSeconds: number = 300) {
        this.outputChannel = outputChannel;
        this.maxPollAttempts = Math.ceil((queryTimeoutSeconds * 1000) / POLL_INTERVAL_MS);

        // Create a reusable insecure agent for servers with certificate issues
        // This mimics how Zowe Explorer handles self-signed/expired certificates
        this.insecureAgent = new https.Agent({
            rejectUnauthorized: false
        });
    }

    private log(message: string): void {
        const debug = vscode.workspace.getConfiguration('smpe').get<boolean>('debug', true);
        if (!debug) { return; }
        const timestamp = new Date().toISOString();
        this.outputChannel.appendLine(`[${timestamp}] [ZosmfClient] ${message}`);
    }

    /**
     * Create Basic Auth header value
     */
    private createAuthHeader(credentials: Credentials): string {
        const encoded = Buffer.from(`${credentials.user}:${credentials.password}`).toString('base64');
        return `Basic ${encoded}`;
    }

    /**
     * Build the CSI query URL
     * Note: xgim.py does NOT include the port in the URL path, only in the connection
     */
    private buildQueryUrl(server: ZosmfServer): string {
        const host = server.host.replace(/\/$/, '');
        const csi = Array.isArray(server.csi) ? server.csi[0] : server.csi;
        const encodedCsi = encodeURIComponent(csi);
        return `${host}/zosmf/swmgmt/csi/csiquery/${encodedCsi}`;
    }

    /**
     * Make an HTTP/HTTPS request
     */
    private async request(
        url: string,
        method: string,
        headers: Record<string, string>,
        body: string | null,
        rejectUnauthorized: boolean
    ): Promise<{ statusCode: number; headers: http.IncomingHttpHeaders; body: string }> {
        return new Promise((resolve, reject) => {
            const parsedUrl = new URL(url);
            const isHttps = parsedUrl.protocol === 'https:';

            // For insecure connections, create a fresh agent with all certificate checks disabled
            let agent: https.Agent | undefined;
            if (isHttps && !rejectUnauthorized) {
                // Temporarily disable NODE_TLS_REJECT_UNAUTHORIZED for this request
                const originalEnv = process.env.NODE_TLS_REJECT_UNAUTHORIZED;
                process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';

                agent = new https.Agent({
                    rejectUnauthorized: false,
                    // Disable all certificate verification
                    checkServerIdentity: () => undefined
                });
                this.log(`Using insecure agent (certificate validation disabled, NODE_TLS_REJECT_UNAUTHORIZED=0)`);

                // Restore after a short delay (request will have started)
                setTimeout(() => {
                    if (originalEnv !== undefined) {
                        process.env.NODE_TLS_REJECT_UNAUTHORIZED = originalEnv;
                    } else {
                        delete process.env.NODE_TLS_REJECT_UNAUTHORIZED;
                    }
                }, 100);
            }

            const options: https.RequestOptions = {
                hostname: parsedUrl.hostname,
                port: parsedUrl.port || (isHttps ? 443 : 80),
                path: parsedUrl.pathname + parsedUrl.search,
                method: method,
                headers: headers,
                agent: agent,
                // Also set rejectUnauthorized directly on options as fallback
                rejectUnauthorized: rejectUnauthorized
            };

            this.log(`Request to ${parsedUrl.hostname}:${options.port}, path=${options.path}, rejectUnauthorized=${rejectUnauthorized}`);

            const requester = isHttps ? https : http;
            const req = requester.request(options, (res) => {
                let data = '';
                res.on('data', (chunk) => {
                    data += chunk;
                });
                res.on('end', () => {
                    resolve({
                        statusCode: res.statusCode || 0,
                        headers: res.headers,
                        body: data
                    });
                });
            });

            req.on('error', (error) => {
                this.log(`Request error: ${error.message}`);
                reject(error);
            });

            if (body) {
                req.write(body);
            }
            req.end();
        });
    }

    /**
     * Execute a SYSMOD query
     * Format from xgim.py:
     * {
     *   "zones": zone,
     *   "entries": ["SYSMOD","TARGETZONE"],
     *   "subentries": ["DELBY,ERROR,FMID,LASTSUP,RECDATE,RECTIME,REWORK,RELATED,SMODTYPE,VERSION,ZONEINDEX"],
     *   "filter": "RELATED!=''|ENAME='sysmod1'|ENAME='sysmod2'"
     * }
     */
    async querySysmod(
        server: ZosmfServer,
        credentials: Credentials,
        zones: string[],
        sysmods: string[],
        progress?: ProgressCallback
    ): Promise<QueryResult> {
        // Build filter: split each entry at spaces/commas to handle list values
        const ids = sysmods.flatMap(sm => sm.split(/[\s,]+/).filter(s => s.length > 0));
        let filterString = "RELATED!=''";
        for (const id of ids) {
            filterString += `|ENAME='${id}'`;
        }

        const body = {
            zones: zones,
            entries: ['SYSMOD', 'TARGETZONE'],
            subentries: ['DELBY,ERROR,FMID,LASTSUP,RECDATE,RECTIME,REWORK,RELATED,SMODTYPE,VERSION,ZONEINDEX'],
            filter: filterString
        };

        return this.executeQuery(server, credentials, body, progress);
    }

    /**
     * Execute a DDDEF query
     * Format from xgim.py:
     * {
     *   "zones": zone,
     *   "entries": ["DDDEF"],
     *   "subentries": ["ENAME,DATASET,DATACLAS,MGMTCLAS,STORCLAS,DIR,DISP,INITDISP,DSNTYPE,SPACE,UNITS,UNIT,VOLUME"],
     *   "filter": "RELATED!=''|ENAME='dddef1'|ENAME='dddef2'"
     * }
     */
    async queryDddef(
        server: ZosmfServer,
        credentials: Credentials,
        zones: string[],
        dddefs: string[],
        progress?: ProgressCallback
    ): Promise<QueryResult> {
        // Build filter string like xgim.py
        let filterString = "RELATED!=''";
        for (const df of dddefs) {
            filterString += `|ENAME='${df}'`;
        }

        const body = {
            zones: zones,
            entries: ['DDDEF'],
            subentries: ['ENAME,DATASET,PATH,DATACLAS,MGMTCLAS,STORCLAS,DIR,DISP,INITDISP,DSNTYPE,SPACE,UNITS,UNIT,VOLUME'],
            filter: filterString
        };

        return this.executeQuery(server, credentials, body, progress);
    }

    /**
     * Execute a zone index query
     * Format from xgim.py:
     * {
     *   "zones": ["GLOBAL"],
     *   "entries": ["GLOBALZONE"],
     *   "subentries": ["ZONEINDEX"],
     *   "filter": "ZONEINDEX!=''"
     * }
     */
    async queryZones(
        server: ZosmfServer,
        credentials: Credentials,
        progress?: ProgressCallback
    ): Promise<QueryResult> {
        const body = {
            zones: ['GLOBAL'],
            entries: ['GLOBALZONE'],
            subentries: ['ZONEINDEX'],
            filter: "ZONEINDEX!=''"
        };

        return this.executeQuery(server, credentials, body, progress);
    }

    /**
     * Execute a free-form CSI query with user-specified parameters
     */
    async queryFreeForm(
        server: ZosmfServer,
        credentials: Credentials,
        zones: string[],
        entryType: string,
        subentries: string[],
        filter: string,
        progress?: ProgressCallback
    ): Promise<QueryResult> {
        // z/OSMF CSI API requires the zone entry type alongside the data entry type
        // to return subentry data. Without it, entries are found but subentries are empty.
        const entries: string[] = [entryType];
        const upperType = entryType.toUpperCase();
        if (upperType !== 'GLOBALZONE' && upperType !== 'TARGETZONE' && upperType !== 'DZONE') {
            // For non-zone entry types, add TARGETZONE so subentries are populated
            entries.push('TARGETZONE');
        }

        const body = {
            zones: zones,
            entries: entries,
            subentries: [subentries.join(',')],
            filter: filter
        };

        return this.executeQuery(server, credentials, body, progress);
    }

    /**
     * Extract error detail from z/OSMF error response body.
     * Per IBM docs: { "error": { "reason": <int>, "messages": ["..."] } }
     * Logs the raw response body and returns a human-readable string.
     */
    private extractZosmfError(rawBody: string): string {
        if (!rawBody) {
            return '(no response body)';
        }
        try {
            const body = JSON.parse(rawBody);
            // IBM documented nested format: { "error": { "reason": <int>, "messages": ["..."] } }
            if (body.error && typeof body.error === 'object') {
                const reason = body.error.reason !== undefined ? ` (reason: ${body.error.reason})` : '';
                const messages = Array.isArray(body.error.messages) && body.error.messages.length > 0
                    ? body.error.messages.join(' | ')
                    : '';
                return messages ? `${messages}${reason}` : `z/OSMF error${reason}`;
            }
            // SMP/E flat format: { "reason": "36", "messages": ["GIM32000W ..."] }
            if (body.messages || body.reason) {
                const reason = body.reason !== undefined ? ` (reason: ${body.reason})` : '';
                const messages = Array.isArray(body.messages) && body.messages.length > 0
                    ? body.messages.join(' | ')
                    : '';
                return messages ? `${messages}${reason}` : `z/OSMF error${reason}`;
            }
            if (body.message) { return body.message; }
        } catch { /* not JSON */ }
        return rawBody.substring(0, 500);
    }

    /**
     * Execute a query and handle async polling
     */
    private async executeQuery(
        server: ZosmfServer,
        credentials: Credentials,
        requestBody: object,
        progress?: ProgressCallback
    ): Promise<QueryResult> {
        const url = this.buildQueryUrl(server);
        this.log(`=== Starting Query ===`);
        this.log(`Server: ${server.name}`);
        this.log(`Host: ${server.host}`);
        this.log(`Port: ${server.port}`);
        this.log(`CSI: ${server.csi}`);
        this.log(`User: ${credentials.user}`);
        this.log(`rejectUnauthorized: ${server.rejectUnauthorized}`);
        this.log(`URL: ${url}`);
        this.log(`Request body: ${JSON.stringify(requestBody)}`);

        const bodyString = JSON.stringify(requestBody);
        const headers: Record<string, string> = {
            'X-CSRF-ZOSMF-HEADER': '',
            'content-type': 'application/json',
            'Content-Length': Buffer.byteLength(bodyString).toString(),
            'Authorization': this.createAuthHeader(credentials)
        };

        progress?.('Sending query to z/OSMF...');

        try {
            const response = await this.request(
                url,
                'POST',
                headers,
                bodyString,
                server.rejectUnauthorized
            );

            this.log(`Response status: ${response.statusCode}`);

            if (response.statusCode === 200) {
                // Synchronous response
                const result = JSON.parse(response.body) as QueryResult;
                this.log('Received synchronous response');
                if (result.entries && result.entries.length > 0) {
                    const firstEntry = result.entries[0];
                    this.log(`First entry keys: ${Object.keys(firstEntry).join(', ')}`);
                    this.log(`First entry subentries type: ${typeof firstEntry.subentries}, isArray: ${Array.isArray(firstEntry.subentries)}, length: ${firstEntry.subentries?.length ?? 'N/A'}`);
                    if (firstEntry.subentries && firstEntry.subentries.length > 0) {
                        this.log(`First subentry sample: ${JSON.stringify(firstEntry.subentries[0])}`);
                    }
                }
                return result;
            } else if (response.statusCode === 202) {
                // Async response - need to poll
                const asyncResponse = JSON.parse(response.body) as AsyncResponse;
                this.log(`Async response, polling: ${asyncResponse.statusurl}`);
                return this.pollForResult(asyncResponse.statusurl, headers, server.rejectUnauthorized, progress);
            } else if (response.statusCode === 400) {
                // The request contained incorrect parameters (e.g. invalid subentry name)
                const detail = this.extractZosmfError(response.body);
                this.log(`HTTP 400 Bad Request. Response: ${response.body}`);
                throw new Error(`HTTP 400 Bad Request: The request contained incorrect parameters. ${detail}`);
            } else if (response.statusCode === 401) {
                this.log(`HTTP 401 Unauthorized. Response: ${response.body}`);
                throw new Error(`HTTP 401 Unauthorized: Authentication failed. Please check your z/OSMF credentials. ${this.extractZosmfError(response.body)}`);
            } else if (response.statusCode === 403) {
                this.log(`HTTP 403 Forbidden. Response: ${response.body}`);
                throw new Error(`HTTP 403 Forbidden: The server rejected the request. ${this.extractZosmfError(response.body)}`);
            } else if (response.statusCode === 404) {
                this.log(`HTTP 404 Not Found. Response: ${response.body}`);
                throw new Error(`HTTP 404 Not Found: The z/OSMF CSI query endpoint was not found. Please verify the host, port, and CSI data set name. ${this.extractZosmfError(response.body)}`);
            } else if (response.statusCode === 409) {
                this.log(`HTTP 409 Conflict. Response: ${response.body}`);
                throw new Error(`HTTP 409 Conflict: The request could not be completed due to a conflict with the current state of the resource. ${this.extractZosmfError(response.body)}`);
            } else if (response.statusCode === 500) {
                this.log(`HTTP 500 Internal Server Error. Response: ${response.body}`);
                throw new Error(`HTTP 500 Internal Server Error: The server encountered an error that prevented it from completing the request. ${this.extractZosmfError(response.body)}`);
            } else if (response.statusCode === 503) {
                this.log(`HTTP 503 Service Unavailable. Response: ${response.body}`);
                throw new Error(`HTTP 503 Service Unavailable: The z/OSMF server is currently unavailable. Please try again later. ${this.extractZosmfError(response.body)}`);
            } else {
                this.log(`HTTP ${response.statusCode}. Response: ${response.body}`);
                throw new Error(`HTTP ${response.statusCode}: ${this.extractZosmfError(response.body)}`);
            }
        } catch (error) {
            if (error instanceof Error) {
                this.log(`Query error: ${error.message}`);
                throw error;
            }
            throw new Error(`Unknown error: ${error}`);
        }
    }

    /**
     * Poll for async query result
     */
    private async pollForResult(
        statusUrl: string,
        headers: Record<string, string>,
        rejectUnauthorized: boolean,
        progress?: ProgressCallback
    ): Promise<QueryResult> {
        let attempts = 0;
        let notFoundCount = 0;

        while (attempts < this.maxPollAttempts) {
            attempts++;
            progress?.(`Waiting for results... (${attempts})`);

            await this.sleep(POLL_INTERVAL_MS);

            try {
                const response = await this.request(
                    statusUrl,
                    'GET',
                    headers,
                    null,
                    rejectUnauthorized
                );

                this.log(`Poll response status: ${response.statusCode}`);

                if (response.statusCode === 200) {
                    const statusResponse = JSON.parse(response.body);
                    this.log(`Poll response body: ${response.body.substring(0, 500)}...`);

                    if (statusResponse.status === 'complete') {
                        this.log('Query completed (async poll)');
                        this.log(`Full async response keys: ${Object.keys(statusResponse).join(', ')}`);
                        // Log first entry's structure for debugging
                        if (statusResponse.entries && statusResponse.entries.length > 0) {
                            const firstEntry = statusResponse.entries[0];
                            this.log(`First entry keys: ${Object.keys(firstEntry).join(', ')}`);
                            this.log(`First entry subentries type: ${typeof firstEntry.subentries}, isArray: ${Array.isArray(firstEntry.subentries)}, length: ${firstEntry.subentries?.length ?? 'N/A'}`);
                            if (firstEntry.subentries && firstEntry.subentries.length > 0) {
                                this.log(`First subentry sample: ${JSON.stringify(firstEntry.subentries[0])}`);
                            }
                        }
                        // z/OSMF returns entries directly in the response, not in a 'result' field
                        // Format: { "status": "complete", "entries": [...] }
                        if (statusResponse.entries) {
                            return { entries: statusResponse.entries } as QueryResult;
                        }
                        return { messages: ['Query completed but no results returned'] };
                    } else if (statusResponse.status === 'failed') {
                        this.log(`Query failed. z/OSMF response: ${response.body}`);
                        throw new Error(`Query failed: ${this.extractZosmfError(response.body)}`);
                    }
                    // status === 'running', continue polling
                } else if (response.statusCode === 202) {
                    // Still processing, continue
                    const asyncResponse = JSON.parse(response.body);
                    if (asyncResponse.statusurl && asyncResponse.statusurl !== statusUrl) {
                        this.log(`Status URL changed to: ${asyncResponse.statusurl}`);
                        return this.pollForResult(asyncResponse.statusurl, headers, rejectUnauthorized, progress);
                    }
                } else if (response.statusCode === 404) {
                    // z/OSMF may return 404 briefly while the query is being initialized
                    notFoundCount++;
                    this.log(`HTTP 404 during poll (retry ${notFoundCount}/${MAX_404_RETRIES}). Response: ${response.body}`);
                    if (notFoundCount >= MAX_404_RETRIES) {
                        throw new Error(`HTTP 404 Not Found: The query status URL was not found after ${MAX_404_RETRIES} retries (${MAX_404_RETRIES * POLL_INTERVAL_MS / 1000}s). The z/OSMF server may be unavailable or the query expired.`);
                    }
                } else if (response.statusCode === 409) {
                    this.log(`HTTP 409 Conflict during poll. Response: ${response.body}`);
                    throw new Error(`HTTP 409 Conflict: The request could not be completed due to a conflict with the current state of the resource. ${this.extractZosmfError(response.body)}`);
                } else if (response.statusCode === 500) {
                    this.log(`HTTP 500 Internal Server Error during poll (attempt ${attempts}). Response: ${response.body}`);
                    // SMP/E application errors (e.g. RC 36 = no entries found) come as HTTP 500
                    // with a flat { "reason": "...", "messages": [...] } body - not transient, throw immediately.
                    try {
                        const errBody = JSON.parse(response.body);
                        if (errBody.reason !== undefined || Array.isArray(errBody.messages)) {
                            throw new Error(this.extractZosmfError(response.body));
                        }
                    } catch (innerErr) {
                        if (innerErr instanceof Error) { throw innerErr; }
                    }
                    // Truly transient server errors: retry up to 3 times
                    if (attempts >= 3) {
                        throw new Error(`HTTP 500 Internal Server Error: The server encountered an error. ${this.extractZosmfError(response.body)}`);
                    }
                } else if (response.statusCode === 503) {
                    this.log(`HTTP 503 Service Unavailable during poll. Response: ${response.body}`);
                    throw new Error(`HTTP 503 Service Unavailable: The z/OSMF server is currently unavailable. Please try again later. ${this.extractZosmfError(response.body)}`);
                } else {
                    this.log(`HTTP ${response.statusCode} during poll. Response: ${response.body}`);
                    throw new Error(`HTTP ${response.statusCode}: ${this.extractZosmfError(response.body)}`);
                }
            } catch (error) {
                if (error instanceof Error && error.message.includes('ECONNRESET')) {
                    // Connection reset, retry
                    this.log('Connection reset, retrying...');
                    continue;
                }
                throw error;
            }
        }

        throw new Error(`Query timed out: no result received after ${this.maxPollAttempts} poll attempts (${this.maxPollAttempts * POLL_INTERVAL_MS / 1000}s).`);
    }

    /**
     * Build the base URL for z/OSMF REST requests (without endpoint path)
     */
    private buildBaseUrl(server: ZosmfServer): string {
        return server.host.replace(/\/$/, '');
    }

    /**
     * List a USS directory via z/OSMF Files REST API
     * GET /zosmf/restfiles/fs?path=<ussPath>
     */
    async listUssDirectory(
        server: ZosmfServer,
        credentials: Credentials,
        ussPath: string
    ): Promise<UssDirectoryListing> {
        const baseUrl = this.buildBaseUrl(server);
        const headers: Record<string, string> = {
            'X-CSRF-ZOSMF-HEADER': '',
            'Authorization': this.createAuthHeader(credentials)
        };

        // Try the full path first; on 404 (reason 8 = path not found),
        // strip the first path segment (likely a PATHPREFIX like /Z31TGT) and retry
        let tryPath = ussPath;
        while (tryPath.length > 1) {
            const url = `${baseUrl}/zosmf/restfiles/fs?path=${tryPath}`;
            this.log(`USS list directory: ${tryPath}`);
            this.log(`USS list URL: ${url}`);

            const response = await this.request(url, 'GET', headers, null, server.rejectUnauthorized);
            this.log(`USS list response status: ${response.statusCode}`);

            if (response.statusCode === 200) {
                const listing = JSON.parse(response.body) as UssDirectoryListing;
                listing.resolvedPath = tryPath;
                return listing;
            }

            this.log(`USS list response body: ${response.body}`);

            // Check if it's a "path not found" error — strip first segment and retry
            if (response.statusCode === 404) {
                try {
                    const errBody = JSON.parse(response.body);
                    if (errBody.reason === 8 || errBody.reason === '8') {
                        // Strip first path segment: /Z31TGT/usr/include → /usr/include
                        const nextSlash = tryPath.indexOf('/', 1);
                        if (nextSlash > 0) {
                            const stripped = tryPath.substring(nextSlash);
                            this.log(`Path not found, stripping prefix: ${tryPath} → ${stripped}`);
                            tryPath = stripped;
                            continue;
                        }
                    }
                } catch { /* not JSON, fall through */ }
            }

            // Non-retryable error
            const detail = this.extractZosmfError(response.body);
            throw new Error(`HTTP ${response.statusCode}: ${detail}`);
        }

        throw new Error(`USS path not found after stripping all segments: ${ussPath}`);
    }

    /**
     * Read a USS file via z/OSMF Files REST API
     * GET /zosmf/restfiles/fs/<ussFilePath>
     */
    async readUssFile(
        server: ZosmfServer,
        credentials: Credentials,
        ussFilePath: string
    ): Promise<string> {
        const baseUrl = this.buildBaseUrl(server);
        const headers: Record<string, string> = {
            'X-CSRF-ZOSMF-HEADER': '',
            'Authorization': this.createAuthHeader(credentials)
        };

        const url = `${baseUrl}/zosmf/restfiles/fs${ussFilePath}`;
        this.log(`USS read file: ${ussFilePath}`);

        const response = await this.request(url, 'GET', headers, null, server.rejectUnauthorized);
        this.log(`USS read response status: ${response.statusCode}`);

        if (response.statusCode === 200) {
            return response.body;
        } else {
            const detail = this.extractZosmfError(response.body);
            throw new Error(`HTTP ${response.statusCode}: ${detail}`);
        }
    }

    /**
     * List PDS members via z/OSMF Dataset REST API
     * GET /zosmf/restfiles/ds/<dataset>/member
     */
    async listDatasetMembers(
        server: ZosmfServer,
        credentials: Credentials,
        datasetName: string
    ): Promise<DatasetMemberListing> {
        const baseUrl = this.buildBaseUrl(server);
        const headers: Record<string, string> = {
            'X-CSRF-ZOSMF-HEADER': '',
            'Authorization': this.createAuthHeader(credentials)
        };

        const url = `${baseUrl}/zosmf/restfiles/ds/${encodeURIComponent(datasetName)}/member`;
        this.log(`Dataset list members: ${datasetName}`);

        const response = await this.request(url, 'GET', headers, null, server.rejectUnauthorized);
        this.log(`Dataset list response status: ${response.statusCode}`);

        if (response.statusCode === 200) {
            return JSON.parse(response.body) as DatasetMemberListing;
        } else {
            const detail = this.extractZosmfError(response.body);
            throw new Error(`HTTP ${response.statusCode}: ${detail}`);
        }
    }

    /**
     * Read a dataset (sequential) or PDS member via z/OSMF Dataset REST API
     * GET /zosmf/restfiles/ds/<dataset>
     * GET /zosmf/restfiles/ds/<dataset>(<member>)
     */
    async readDataset(
        server: ZosmfServer,
        credentials: Credentials,
        datasetName: string,
        memberName?: string
    ): Promise<string> {
        const baseUrl = this.buildBaseUrl(server);
        const headers: Record<string, string> = {
            'X-CSRF-ZOSMF-HEADER': '',
            'Authorization': this.createAuthHeader(credentials)
        };

        const dsn = memberName
            ? `${datasetName}(${memberName})`
            : datasetName;
        const url = `${baseUrl}/zosmf/restfiles/ds/${encodeURIComponent(dsn)}`;
        this.log(`Dataset read: ${dsn}`);

        const response = await this.request(url, 'GET', headers, null, server.rejectUnauthorized);
        this.log(`Dataset read response status: ${response.statusCode}`);

        if (response.statusCode === 200) {
            return response.body;
        } else {
            const detail = this.extractZosmfError(response.body);
            throw new Error(`HTTP ${response.statusCode}: ${detail}`);
        }
    }

    /**
     * Sleep for specified milliseconds
     */
    private sleep(ms: number): Promise<void> {
        return new Promise(resolve => setTimeout(resolve, ms));
    }
}
