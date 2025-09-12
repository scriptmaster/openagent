declare interface Page {
    PageTitle: string;
    Content: string;
    AppName: string;
    AppVersion: string;
    AdminEmail: string;

    User: AuthUser;

    Error: string;

}

declare interface AuthUser {
    ID:           number;
	Email:        string;
	IsAdmin:      boolean;
	MemberSince:  Date; // CreatedAt
	LastLoggedIn: Date;
}
