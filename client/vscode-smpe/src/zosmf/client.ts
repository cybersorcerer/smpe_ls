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
    ProgressCallback
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
        const encodedCsi = encodeURIComponent(server.csi);
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

            this.log(`Request to ${parsedUrl.hostname}:${options.port}, rejectUnauthorized=${rejectUnauthorized}`);

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
        // Build filter string like xgim.py
        let filterString = "RELATED!=''";
        for (const sm of sysmods) {
            filterString += `|ENAME='${sm}'`;
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
            subentries: ['ENAME,DATASET,DATACLAS,MGMTCLAS,STORCLAS,DIR,DISP,INITDISP,DSNTYPE,SPACE,UNITS,UNIT,VOLUME'],
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
        const body = {
            zones: zones,
            entries: [entryType],
            subentries: [subentries.join(',')],
            filter: filter
        };

        return this.executeQuery(server, credentials, body, progress);
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
                return result;
            } else if (response.statusCode === 202) {
                // Async response - need to poll
                const asyncResponse = JSON.parse(response.body) as AsyncResponse;
                this.log(`Async response, polling: ${asyncResponse.statusurl}`);
                return this.pollForResult(asyncResponse.statusurl, headers, server.rejectUnauthorized, progress);
            } else if (response.statusCode === 401) {
                this.log(`HTTP 401 Unauthorized. Response: ${response.body}`);
                throw new Error('HTTP 401 Unauthorized: Authentication failed. Please check your z/OSMF credentials.');
            } else if (response.statusCode === 403) {
                this.log(`HTTP 403 Forbidden. Response: ${response.body}`);
                throw new Error('HTTP 403 Forbidden: Access denied. The user may not have permission to access this CSI data set.');
            } else if (response.statusCode === 404) {
                this.log(`HTTP 404 Not Found. Response: ${response.body}`);
                throw new Error(`HTTP 404 Not Found: The z/OSMF CSI query endpoint was not found. Please verify the host, port, and CSI data set name.`);
            } else {
                this.log(`HTTP ${response.statusCode}. Response: ${response.body}`);
                let detail = response.body ? response.body.substring(0, 500) : '(no response body)';
                try {
                    const errorBody = JSON.parse(response.body);
                    if (errorBody.message) {
                        detail = errorBody.message;
                    } else if (errorBody.error) {
                        detail = errorBody.error;
                    } else if (errorBody.reason) {
                        detail = errorBody.reason;
                    }
                } catch { /* use raw body */ }
                throw new Error(`HTTP ${response.statusCode}: ${detail}`);
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
                        this.log('Query completed');
                        // z/OSMF returns entries directly in the response, not in a 'result' field
                        // Format: { "status": "complete", "entries": [...] }
                        if (statusResponse.entries) {
                            return { entries: statusResponse.entries } as QueryResult;
                        }
                        return { messages: ['Query completed but no results returned'] };
                    } else if (statusResponse.status === 'failed') {
                        const reason = statusResponse.error || statusResponse.reason || statusResponse.message || '(no details provided)';
                        this.log(`Query failed. z/OSMF response: ${response.body}`);
                        throw new Error(`Query failed: ${reason}`);
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
                } else if (response.statusCode === 500) {
                    this.log(`HTTP 500 during poll (attempt ${attempts}). Response: ${response.body}`);
                    if (attempts >= 3) {
                        let detail = response.body ? response.body.substring(0, 500) : '(no response body)';
                        try {
                            const errorBody = JSON.parse(response.body);
                            detail = errorBody.message || errorBody.reason || detail;
                        } catch { /* use raw body */ }
                        throw new Error(`HTTP 500 Internal Server Error: ${detail}`);
                    }
                    // Retry on first 2 occurrences
                } else {
                    this.log(`HTTP ${response.statusCode} during poll. Response: ${response.body}`);
                    throw new Error(`HTTP ${response.statusCode}: Unexpected response while polling for query results.`);
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
     * Sleep for specified milliseconds
     */
    private sleep(ms: number): Promise<void> {
        return new Promise(resolve => setTimeout(resolve, ms));
    }
}
