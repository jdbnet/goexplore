export namespace config {
	
	export class ConnectionConfig {
	    id: string;
	    name: string;
	    protocol: string;
	    host?: string;
	    port?: number;
	    bucket?: string;
	    region?: string;
	    path_style?: boolean;
	    secure?: boolean;
	    username?: string;
	    keychain_key?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.protocol = source["protocol"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.bucket = source["bucket"];
	        this.region = source["region"];
	        this.path_style = source["path_style"];
	        this.secure = source["secure"];
	        this.username = source["username"];
	        this.keychain_key = source["keychain_key"];
	    }
	}

}

export namespace explorer {
	
	export class FileEntry {
	    name: string;
	    path: string;
	    size: number;
	    modified: string;
	    is_dir: boolean;
	    permissions: string;
	
	    static createFrom(source: any = {}) {
	        return new FileEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	        this.modified = source["modified"];
	        this.is_dir = source["is_dir"];
	        this.permissions = source["permissions"];
	    }
	}

}

export namespace main {
	
	export class TransferItem {
	    path: string;
	    name: string;
	    is_dir: boolean;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new TransferItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.is_dir = source["is_dir"];
	        this.size = source["size"];
	    }
	}

}

export namespace transfer {
	
	export class Transfer {
	    id: string;
	    source: string;
	    destination: string;
	    filename: string;
	    bytes_total: number;
	    bytes_done: number;
	    speed_mbps: number;
	    eta_seconds: number;
	    status: string;
	    error?: string;
	    verify: boolean;
	    limit_mbps: number;
	
	    static createFrom(source: any = {}) {
	        return new Transfer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.source = source["source"];
	        this.destination = source["destination"];
	        this.filename = source["filename"];
	        this.bytes_total = source["bytes_total"];
	        this.bytes_done = source["bytes_done"];
	        this.speed_mbps = source["speed_mbps"];
	        this.eta_seconds = source["eta_seconds"];
	        this.status = source["status"];
	        this.error = source["error"];
	        this.verify = source["verify"];
	        this.limit_mbps = source["limit_mbps"];
	    }
	}

}

