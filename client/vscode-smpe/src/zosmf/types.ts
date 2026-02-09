/**
 * z/OSMF SMP/E Query Integration Types
 */

// ============================================================================
// Configuration Types
// ============================================================================

/**
 * Single z/OSMF server configuration
 */
export interface ZosmfServer {
    name: string;
    host: string;
    port: number;
    csi: string;
    user: string;
    rejectUnauthorized: boolean;
    defaultZones?: string[];
}

/**
 * Root configuration structure from .smpe-zosmf.yaml
 */
export interface ZosmfConfig {
    servers: ZosmfServer[];
    defaultServer?: string;
}

// ============================================================================
// API Request Types
// ============================================================================

/**
 * Query types supported by z/OSMF SMP/E API
 */
export type QueryType = 'sysmod' | 'dddef' | 'zone';

/**
 * SYSMOD query request body
 */
export interface SysmodQueryRequest {
    zones: string[];
    sysmods: string[];
    subentries?: string[];
}

/**
 * DDDEF query request body
 */
export interface DddefQueryRequest {
    zones: string[];
    dddefs: string[];
}

/**
 * Zone index query request body
 */
export interface ZoneIndexQueryRequest {
    zones?: string[];
}

/**
 * Union type for all query request bodies
 */
export type QueryRequest = SysmodQueryRequest | DddefQueryRequest | ZoneIndexQueryRequest;

// ============================================================================
// API Response Types
// ============================================================================

/**
 * z/OSMF async response (HTTP 202)
 */
export interface AsyncResponse {
    statusurl: string;
}

/**
 * z/OSMF status poll response
 */
export interface StatusResponse {
    status: 'running' | 'complete' | 'failed';
    statusurl?: string;
    result?: QueryResult;
    error?: string;
}

/**
 * z/OSMF CSI Query Entry (raw format from API)
 * Example: { "entryname": "HZDC7C0", "entrytype": "SYSMOD", "zonename": "MVST100", "subentries": [...] }
 */
export interface ZosmfEntry {
    entryname: string;
    entrytype: string;
    zonename: string;
    subentries: ZosmfSubentry[];
}

/**
 * z/OSMF Subentry (key-value pairs with VER field)
 * Example: { "FMID": ["HZDC7C0"], "VER": null }
 */
export interface ZosmfSubentry {
    [key: string]: string[] | null;
}

/**
 * Query result containing raw z/OSMF entries
 */
export interface QueryResult {
    entries?: ZosmfEntry[];
    messages?: string[];
}

// ============================================================================
// Internal Types
// ============================================================================

/**
 * Authentication credentials
 */
export interface Credentials {
    user: string;
    password: string;
}

/**
 * Query execution context
 */
export interface QueryContext {
    server: ZosmfServer;
    credentials: Credentials;
    queryType: QueryType;
}

/**
 * Query progress callback
 */
export type ProgressCallback = (message: string) => void;

/**
 * Query result for display
 */
export interface DisplayResult {
    serverName: string;
    queryType: QueryType;
    timestamp: Date;
    result: QueryResult;
    error?: string;
}
