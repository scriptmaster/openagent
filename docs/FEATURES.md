1. Login:
    Uses JWT
    Uses OTP as the primary authentication system
2. Projects:
    Each project is loaded and cached by host or domain name from the request.
    Each project can have its own database, again the connection can be cached [?]
    In the project configuration screen. Options for:
        A) Change database name or schema (default schema: public), (with the current connection string and db name)
        B) Change only schema name (with the current connection string)
        C) Change the entire database connection: driver (postgres, mssql, mysql), connection string.
        D) resulting string store in options. connection string stored in the project table options column.
        E) dont expose the current connection string on UI for A) and B). Get the inputs and handle in the backend.
3. Users:
    If no project, Users are loaded from the default ai.users table for login as admin.
    If a project is configured then users are loaded from its connecting database. from project.options
    Users are authenticated with the standard /login page.
    No registration page. If a user logs in for the first time, their account is created., with the OTP flow.
    A profile page lets them specify their Name (defaults to their email), set their password with OTP verification for password change. (password change is a seperate section or tabbed interface in the profile page)
    Set a gravatar or profile icon.

PLANNED FEATURES (Don't implement them yet):
4. Agent
    with the /agent route. Ensure it is accessible and loading the right template.

5. Voice
    with the /voice route. Ensure it is accessible and loading the right template.
