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
    csi: string | string[];
    defaultCsi?: string;
    user: string;
    rejectUnauthorized: boolean;
    zones?: string[];
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
    requestedIds?: string[];
}

// ============================================================================
// USS File Types (z/OSMF Files REST API)
// ============================================================================

/**
 * Single USS directory entry from z/OSMF Files REST API
 * GET /zosmf/restfiles/fs?path=...
 */
export interface UssEntry {
    name: string;
    mode: string;
    size: number;
    uid: number;
    user: string;
    gid: number;
    group: string;
    mtime: string;
}

/**
 * USS directory listing response
 */
export interface UssDirectoryListing {
    items: UssEntry[];
    returnedRows: number;
    totalRows: number;
    JSONversion: number;
    /** The actual resolved path (after PATHPREFIX stripping) */
    resolvedPath?: string;
}

// ============================================================================
// MVS Dataset Types (z/OSMF Files REST API)
// ============================================================================

/**
 * Single PDS member entry from z/OSMF Dataset REST API
 * GET /zosmf/restfiles/ds/<dataset>/member
 */
export interface DatasetMember {
    member: string;
    vers?: number;
    mod?: number;
    c4date?: string;
    m4date?: string;
    cnorc?: number;
    inorc?: number;
    mnorc?: number;
    mtime?: string;
    msec?: string;
    user?: string;
    sclm?: string;
}

/**
 * PDS member listing response
 */
export interface DatasetMemberListing {
    items: DatasetMember[];
    returnedRows: number;
    totalRows: number;
    JSONversion: number;
}
