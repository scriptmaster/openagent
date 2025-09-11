declare interface Page {
    PageTitle: string;
    Content: string;
    AppName: string;
    AppVersion: string;

    User: AuthUser;
}

declare interface AuthUser {
    ID:           number;
	Email:        string;
	IsAdmin:      boolean;
	MemberSince:  Date; // CreatedAt
	LastLoggedIn: Date;
}
