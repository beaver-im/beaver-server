# 🤝 Contributing to Beaver IM

Thank you for your interest in contributing to Beaver IM! This document provides guidelines and information for contributors.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Coding Standards](#coding-standards)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Reporting Bugs](#reporting-bugs)
- [Feature Requests](#feature-requests)
- [Documentation](#documentation)

## 📜 Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

### Our Standards

- **Be respectful** - Treat everyone with respect
- **Be collaborative** - Work together to achieve common goals
- **Be constructive** - Provide helpful feedback and suggestions
- **Be professional** - Maintain professional behavior in all interactions

## 🎯 How Can I Contribute?

### 🐛 Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates.

**Bug Report Template:**

```markdown
## Bug Description
Brief description of the bug

## Steps to Reproduce
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

## Expected Behavior
What you expected to happen

## Actual Behavior
What actually happened

## Environment
- OS: [e.g. Windows 10, macOS 12.0]
- Go Version: [e.g. 1.26.8]
- Database: [e.g. MySQL 8.0]
- Redis: [e.g. 6.0]

## Additional Information
Any additional context, logs, or screenshots
```

### 💡 Suggesting Enhancements

**Feature Request Template:**

```markdown
## Problem Statement
Clear description of the problem this feature would solve

## Proposed Solution
Description of the proposed solution

## Alternative Solutions
Any alternative solutions you've considered

## Additional Context
Any other context, screenshots, or examples
```

### 🔧 Code Contributions

We welcome code contributions! Here's how to get started:

1. **Fork** the repository
2. **Create** a feature branch
3. **Make** your changes
4. **Test** your changes
5. **Submit** a pull request

## 🛠️ Development Setup

### Prerequisites

- Go >= 1.26（以 [go.mod](go.mod) 为准，当前为 1.26.8）
- MySQL >= 8.0
- Redis >= 6.0
- ETCD >= 3.5
- Docker >= 20.0

### Local Development

1. **Clone your fork**
```bash
git clone https://github.com/YOUR_USERNAME/beaver-server.git
cd beaver-server
```

2. **Add upstream remote**
```bash
git remote add upstream https://github.com/wsrh8888/beaver-server.git
```

3. **Install dependencies**
```bash
go mod tidy
```

4. **Start development environment**
```bash
# Start infrastructure services（事实源：docker/docker-compose.yml，含 MySQL / Redis / ETCD / MQ 等）
docker-compose -f docker/docker-compose.yml up -d

# 本地开发建库辅助（不是总服务启动器，也不是生产共库部署）
go run main.go -db
```

5. **Run tests**
```bash
go test ./...
```

## 📝 Coding Standards

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for code formatting
- Use `go vet ./...` for static checks（本仓未配置 `golint`，不再推荐）
- Follow naming conventions:
  - `camelCase` for variables and functions
  - `PascalCase` for exported names
  - `snake_case` for file names

### Project Structure

```
app/
└── <domain>/                       # 业务域，如 account / message / relation
    ├── <domain>_api/               # HTTP 服务进程
    │   ├── api/                    # .api 契约（事实源，goctl 生成）
    │   ├── internal/
    │   │   ├── config/             # 配置结构
    │   │   ├── handler/            # HTTP handler（goctl 生成）
    │   │   ├── logic/              # 业务逻辑扩展点
    │   │   ├── middleware/         # 中间件
    │   │   ├── svc/                # ServiceContext
    │   │   └── types/              # 请求/响应类型（goctl 生成）
    │   └── etc/                    # 运行配置
    ├── <domain>_rpc/               # RPC 服务进程
    │   ├── internal/{config,logic,server,svc}/
    │   ├── client/                 # 本仓跨域 RPC 客户端
    │   ├── pb/                     # goctl 生成的 pb
    │   └── etc/
    └── <domain>_models/            # 本域独占库的数据模型（不得跨库引用）
```

各域独占数据库、物理隔离部署；跨域调用走 RPC 或 remote，不连对方库。
目录细则见 [工程目录](docs/directory.md)。

### Error Handling

- Always check errors and handle them appropriately
- Use custom error types for business logic errors
- Provide meaningful error messages
- Log errors with appropriate levels

### Logging

业务日志统一走 `utils/beaverlog` 包装器，不直接用 `logx`。logger 挂在 logic 上，构造时传入 ctx：

```go
import (
	"context"

	beaverlog "beaver/utils/beaverlog"
	"beaver/utils/beaverlog/model"
)

type AuthTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logger *beaverlog.Logger
}

func NewAuthTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthTokenLogic {
	return &AuthTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		logger: beaverlog.New("auth_token", ctx),
	}
}

// Text 为中文短语（不带空格）；结构化字段放 Data。
// user_id / traceId 等公参由 logger 从 ctx 自动抽取，不要手写进日志。
l.logger.Error(model.LogMsg{
	Text: "秘钥签发被拒",
	Data: map[string]any{"clientID": req.ClientID},
})
```

API 为 `logger.Info / Warn / Error(model.LogMsg{Text, Data})`。禁止把令牌、请求头、完整聊天内容或
未脱敏配置写进日志。细则见 [5.4 错误与日志](docs/05-runtime/04-错误与日志.md)。

### Testing

- Write unit tests for all business logic
- Aim for >80% code coverage
- Use table-driven tests for multiple scenarios
- Mock external dependencies

```go
func TestUserService_Register(t *testing.T) {
    tests := []struct {
        name    string
        input   RegisterRequest
        want    *RegisterResponse
        wantErr bool
    }{
        // Test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## 📝 Commit Guidelines

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Commit Message Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `perf`: A code change that improves performance
- `test`: Adding missing tests or correcting existing tests
- `chore`: Changes to the build process or auxiliary tools

### Examples

```bash
feat(auth): add email verification functionality
fix(chat): resolve message delivery issue
docs: update API documentation
refactor(user): simplify user registration logic
test(api): add integration tests for user endpoints
```

## 🔄 Pull Request Process

### Before Submitting

1. **Update your branch**
```bash
git fetch upstream
git rebase upstream/main
```

2. **Run tests**
```bash
go test ./...
go vet ./...
```

测试以本仓实际存在的 `_test.go` 为准，不虚构覆盖；批量生成、构建、部署需按任务授权执行，
验证口径见 [8.2 验证矩阵](docs/08-workflow/02-验证矩阵.md)。

3. **Check formatting**
```bash
gofmt -s -w .
```

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] My code follows the style guidelines of this project
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
```

### Review Process

1. **Automated Checks**
   - CI/CD pipeline runs tests
   - Code coverage is checked
   - Linting and formatting are verified

2. **Code Review**
   - At least one maintainer must approve
   - Address all review comments
   - Update documentation if needed

3. **Merge**
   - Squash commits if requested
   - Use conventional commit message
   - Delete feature branch after merge

## 📚 Documentation

### Code Documentation

- Document all exported functions and types
- Use clear and concise comments
- Include examples for complex functions

```go
// UserService handles user-related operations
type UserService struct {
    // ...
}

// Register creates a new user account
// Returns the created user ID and any validation errors
func (s *UserService) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
    // Implementation
}
```

### API Documentation

- Update API documentation for new endpoints
- Include request/response examples
- Document error codes and messages

### README Updates

- Update README.md for new features
- Add installation instructions for new dependencies
- Update configuration examples

## 🏷️ Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):

- `MAJOR.MINOR.PATCH`
- `MAJOR`: Breaking changes
- `MINOR`: New features (backward compatible)
- `PATCH`: Bug fixes (backward compatible)

### Release Checklist

- [ ] All tests pass
- [ ] Documentation is updated
- [ ] Changelog is updated
- [ ] Version is bumped
- [ ] Release notes are written
- [ ] Docker images are built and pushed

## 🆘 Getting Help

- **Issues**: [GitHub Issues](https://github.com/wsrh8888/beaver-server/issues)
- **Discussions**: [GitHub Discussions](https://github.com/wsrh8888/beaver-server/discussions)
- **Email**: [751135385@qq.com](mailto:751135385@qq.com)
- **QQ Group**: [1013328597](https://qm.qq.com/q/82rbf7QBzO)

## 🙏 Recognition

Contributors will be recognized in:

- Project README.md
- Release notes
- Contributor hall of fame
- GitHub contributors page

---

Thank you for contributing to Beaver IM! 🦫

## 本仓当前工程规范

上文贡献流程通用；涉及契约、目录、命令与工具版本时，以 [中文工程手册](docs/README.md)、
[工程目录](docs/directory.md) 与 [实施/验证清单](docs/08-workflow/README.md) 及真实源码为准。
`.api` / `.proto` 是接口事实源，goctl 生成接线；生成、构建、启动、部署须在任务授权范围内。
