# Engineering Conventions

## Conventions
>
> When writing code, follow these conventions.

- Write simple, verbose code over terse, compact, dense code.
- Do test driven development, write tests before writing code.
- When the test is written, write the code to pass the test.
- If a function does not have a corresponding test, mention it.
- When building tests, don't mock anything.

## Project Structure

The project is structured according to conventional Go project structure.
It also follows Cobra and Viper conventions.

## Cobra and Viper Conventions

- **Command Structure**: Organize commands in a modular fashion. Each command should be self-contained, managing its own configurations and defaults.
- **Configuration Management**: Use Viper to handle configuration files, environment variables, and command-line flags. Ensure that the order of precedence is: Flags > Environment Variables > Config Files > Defaults.
- **Command Initialization**: Initialize Cobra commands in a `cmd` directory. Each command should be defined in its own file for clarity and maintainability.
- **Error Handling**: Use Cobra's built-in error handling mechanisms to provide meaningful error messages and exit codes.
- **Documentation**: Document each command with a detailed description and usage examples. Use Cobra's built-in help generation features.

## Zerolog Conventions

- **Structured Logging**: Use structured logging to provide context-rich log messages. This helps in filtering and searching logs effectively.
- **Centralized Initialization**: Initialize the logger in a centralized location, such as `internal/logger/logger.go`, to ensure consistent logging configuration across the application.
- **Log Levels**: Use appropriate log levels (e.g., Debug, Info, Warn, Error) to categorize log messages. Avoid using global variables for loggers; pass logger instances where needed.
- **Error Logging**: Log errors with contextual information using `log.Error().Err(err).Msg("error message")`. This provides a clear understanding of the error context.
- **Output Flexibility**: Accept an `io.Writer` parameter in logger initialization to allow flexibility in directing log outputs, which is useful for testing.
