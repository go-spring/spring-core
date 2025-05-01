# Contributing to Go-Spring

First of all, thank you for your interest in and support of the Go-Spring project!

We welcome all kinds of contributions, including reporting issues, improving documentation, fixing bugs, and developing
new features. Please follow the guidelines below to contribute:

## Submitting Issues

- Before submitting, please search existing issues to avoid duplicates.
- Provide clear reproduction steps, expected behavior, and actual results.
- If available, include error logs and relevant environment information.

## Submitting Pull Requests

1. **Fork the repository and create a new branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Maintain consistent coding style**
    - Follow Go’s official style guidelines (use `gofmt`, `golint`, `go vet`)
    - It’s recommended to use [`golangci-lint`](https://github.com/golangci/golangci-lint) for local linting

3. **Write tests**
    - All new features must include unit tests
    - Use Go’s built-in `testing` package, and name test files as `xxx_test.go`

4. **Update documentation (if applicable)**

5. **Submit and create a Pull Request**
    - Clearly describe the purpose, changes made, and testing results
    - Link relevant issues (if any)

## Local Development Environment

- Go version: Latest stable release is recommended (e.g., `go1.21+`)
- Use Go Modules for dependency management
- Make sure all tests pass:
  ```bash
  go test ./...
  ```

## Contact Us

If you have any questions, feel free to open an issue or join the discussion forum.

Thank you for contributing!

--- 

# Contributing to Go-Spring

首先，感谢你关注并支持 Go-Spring 项目！

我们欢迎各种形式的贡献，包括但不限于 Issue 提交、文档完善、Bug 修复、功能开发等。请按照以下指引参与贡献：

## 提交 Issue

- 在提交前，请先搜索现有的 Issue，避免重复提交。
- 请提供清晰的复现步骤、预期行为以及实际结果。
- 如有错误日志或运行环境信息，请一并附上。

## 提交 Pull Request

1. **Fork 仓库并创建新分支**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **保持一致的代码风格**
    - 遵循 Go 官方代码规范（使用 `gofmt`、`golint`、`go vet`）
    - 推荐使用 [`golangci-lint`](https://github.com/golangci/golangci-lint) 进行本地代码检查

3. **编写测试用例**
    - 所有新功能必须配备单元测试
    - 使用 Go 内置的 `testing` 包，测试文件应命名为 `xxx_test.go`

4. **更新相关文档（如有变更）**

5. **提交并创建 Pull Request**
    - 说明 PR 的目的、变更内容、测试情况等
    - 关联相关 Issue（如有）

## 本地开发环境要求

- Go 版本：推荐使用最新版稳定版（如 `go1.21+`）
- 使用 Go Modules 进行依赖管理
- 确保测试全部通过：
  ```bash
  go test ./...
  ```

## 联系我们

如有疑问，欢迎通过 Issue 与我们联系，或参与项目的讨论区。

感谢你的贡献！
