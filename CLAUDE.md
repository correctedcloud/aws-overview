# CLAUDE.md - AWS Overview Project

## Build & Test Commands
- Build: `make build`
- Run tests: `make test`
- Run single test: `go test -v github.com/correctedcloud/aws-overview/pkg/alb -run TestGetLoadBalancers`
- Run with coverage: `go test -cover ./...`
- You can check command output by running `make build && ./aws-overview`

## Code Style Guidelines
- **Formatting**: Use `gofmt` or `go fmt ./...` before committing
- **Imports**: Group standard library imports first, then external dependencies, then internal packages
- **Types**: Use strong typing; prefer explicit types over `interface{}`
- **Error Handling**: Return errors rather than using panics; check all errors
- **Naming**: Use CamelCase for exported names, camelCase for unexported names
- **Testing**: Write unit tests for all packages; use table-driven tests when appropriate
- **Comments**: Add comments for all exported functions, types, and constants
- **Dependencies**: Import aliases for clarity when names conflict (e.g., `rdssvc`)
- **Context**: Pass context through function calls for AWS SDK operations
- **Function Size**: Keep functions small and focused on a single task

## AWS SDK Usage
- Use AWS SDK Go v2 for all AWS service interactions
- Use proper error handling for all AWS API calls
- Follow the client pattern used in alb/rds packages
