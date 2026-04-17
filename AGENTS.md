# AGENTS.md

## Coding Conventions

Any code changes must pass a `make lint` without any errors

This uses Go 1.25+, and should use the more modern conventions these later versions provide.

Unit testing should test critical functionality, not necessarily aim for 100% coverage, to keep tests easy to maintain.
I want high confidence the code will work, but do not want to spend much time refactoring unit tests as I make changes.
Use the testify package when writing unit tests.

