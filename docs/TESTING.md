1. Basic UI/UX Testing:
    For UI testing we will not use playwright in basic testing.
    These tests will be in tests/seo/test_name.go files
    They will be curl/http commands in go with one or two host names specified including a third one for localhost
    The testing environment can be either localhost, a staging host or a production host.
    Multiple testing environments can be configured via a .env.name files.
    Testing setup will loop and rerun for all envs.
    http result from the root pages to project functionalities will test the page response status, response html.
    html parsing can be used. meta tags will be tested. page content and ui tags, h1 and h2 presence will be checked.
    http headers will be checked. CSP, HSTS, other security headers will be checked.
    HTML Content will be checked for matching content for each page.
    User login will be tested., if they are in the right databse.
    Since OTP testing cannot be done, user will be chosen to test with their password only.

2. API Testing:
    Generate tests from api docs available from the openapi specification of the project.
    both authenticated and unauthenticated api endpoints will be tested.

3. Code Coverage
    Run make test to ensure all tests are run. Add any missing tests. Run the tests. Ensure testing code coverage is 100%.
