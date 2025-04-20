Sometimes the tool call limit of 25 is reached or simply the agent stops in middle of implementation, so copy paste these prompts wherever needed. Press Ctrl+Shift+V if you are pasting into cursor agent.

-- PROMPT: FEATURE --
Implement the missing features from docs/FEATURES.md

also correct the errors from go vet ./...

Ensure before and after implementation of the features, git add, commit, push and go vet ./... commands are run in a loop to ensure there are no errors. Keep doing until done.

-- PROMPT: TESTING --
Run make test to ensure all tests are run. Add any missing tests. Run the tests. Ensure testing code coverage is 100%.
Then Run "make" command for starting the server and check if any errors. Ensure and loop until server has successfully and run for 3 seconds.
Then Run the ui tests for basic testing. If ui tests are not created, read docs/TESTING.md for how to create basic UI tests.

-- PROMPT: API --
Run the api tests to ensure all api testing are done. For all database operations there should be an api endpoint. if not create it, in a new api/routes.go file.
