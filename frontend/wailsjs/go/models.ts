export namespace app {
	
	export class MessageSummaryDTO {
	    uid: number;
	    subject: string;
	    from: string;
	    date: string;
	    dateUnix: number;
	    seen: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MessageSummaryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.uid = source["uid"];
	        this.subject = source["subject"];
	        this.from = source["from"];
	        this.date = source["date"];
	        this.dateUnix = source["dateUnix"];
	        this.seen = source["seen"];
	    }
	}
	export class FolderDTO {
	    name: string;
	    displayName: string;
	    delimiter: string;
	
	    static createFrom(source: any = {}) {
	        return new FolderDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.displayName = source["displayName"];
	        this.delimiter = source["delimiter"];
	    }
	}
	export class AccountCacheDTO {
	    folders: FolderDTO[];
	    lastFolder: string;
	    messages: MessageSummaryDTO[];
	
	    static createFrom(source: any = {}) {
	        return new AccountCacheDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.folders = this.convertValues(source["folders"], FolderDTO);
	        this.lastFolder = source["lastFolder"];
	        this.messages = this.convertValues(source["messages"], MessageSummaryDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AccountDTO {
	    id: string;
	    email: string;
	    label: string;
	    provider: string;
	    providerTag: string;
	    isActive: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AccountDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.email = source["email"];
	        this.label = source["label"];
	        this.provider = source["provider"];
	        this.providerTag = source["providerTag"];
	        this.isActive = source["isActive"];
	    }
	}
	export class AttachmentMetaDTO {
	    filename: string;
	    contentType: string;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new AttachmentMetaDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filename = source["filename"];
	        this.contentType = source["contentType"];
	        this.size = source["size"];
	    }
	}
	
	export class LoginResult {
	    success: boolean;
	    email: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new LoginResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.email = source["email"];
	        this.error = source["error"];
	    }
	}
	export class MessageDetailDTO {
	    uid: number;
	    subject: string;
	    from: string;
	    to: string[];
	    date: string;
	    textPlain: string;
	    textHtml: string;
	    attachments: AttachmentMetaDTO[];
	
	    static createFrom(source: any = {}) {
	        return new MessageDetailDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.uid = source["uid"];
	        this.subject = source["subject"];
	        this.from = source["from"];
	        this.to = source["to"];
	        this.date = source["date"];
	        this.textPlain = source["textPlain"];
	        this.textHtml = source["textHtml"];
	        this.attachments = this.convertValues(source["attachments"], AttachmentMetaDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class ProviderDTO {
	    id: string;
	    displayName: string;
	    authType: string;
	    helpUrl: string;
	    emailPlaceholder: string;
	
	    static createFrom(source: any = {}) {
	        return new ProviderDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.displayName = source["displayName"];
	        this.authType = source["authType"];
	        this.helpUrl = source["helpUrl"];
	        this.emailPlaceholder = source["emailPlaceholder"];
	    }
	}
	export class SettingsDTO {
	    fetchDays: number;
	    refreshIntervalMinutes: number;
	    openLinksInBrowser: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SettingsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fetchDays = source["fetchDays"];
	        this.refreshIntervalMinutes = source["refreshIntervalMinutes"];
	        this.openLinksInBrowser = source["openLinksInBrowser"];
	    }
	}

}

