# Contributing to PomboHook

Thank you for your interest in contributing to **PomboHook**! 🎉

This document serves as a guide to help you understand our development process, required coding standards, and how you can submit your changes.

## 🤝 Submission Process

Before you start writing code for a new feature or a bug fix, **you must open an Issue**.

1. **Open an Issue:** Go to the "Issues" tab on GitHub and describe in detail what you want to implement or what problem you've found.
2. **Discussion:** Wait for validation from the community or the maintainer. This prevents you from wasting time implementing something that might not align with the project's vision or is already being worked on by someone else.
3. **Development:** After the Issue is approved, you can fork the repository, create your branch, and start coding.
4. **Pull Request:** Open your PR referencing the original Issue (e.g., `Resolves #12`).

## 💻 Code and Engineering Standards

To keep the project sustainable and organized, we follow strict engineering rules. When contributing, please make sure to follow these guidelines:

1. **Simplicity:** Prefer small functions and files. Avoid deep nesting (use *early returns* and *guard clauses*).
2. **Single Responsibility (SRP):** Each module, class, and function should have only one responsibility.
3. **Naming:** Use names that reveal intent (avoid generic names like `data`, `process`, or `handler`).
4. **Comments:** Do not comment on what is obvious. Use comments only to explain non-trivial decisions or bug workarounds.
5. **Actionable Errors:** Error messages and exceptions must include enough context to identify the problem quickly. Do not use vague messages.
6. **Logging:** Use structured logging for observability and plain text logs for user-facing output in the CLI.

## 🧪 Testing Workflow

PomboHook takes testing very seriously. For your Pull Request to be accepted, the following test requirements **must be strictly met**:

1. **Local Execution:** The commands `make test` and `make lint` (or `go vet ./...`) must pass without any errors on your machine.
2. **TDD and Behavior:** Test public behavior, *edge cases*, and failure paths.
3. **Coverage:** 
   - New code requires **100% coverage** for new lines.
   - The project has a global floor of **80% coverage**, and critical modules (auth, routing, tunnel) require at least **90%**.
4. **Mocking:** Do not mock the unit under test, mock only its dependencies.

## 📝 Commit Standards

We adopt the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) convention. Your commit messages must follow this standard format:

```text
<type>[optional scope]: <description>
```

**Valid examples:**
- `feat: add persistent sqlite storage`
- `fix: resolve data race in tunnel manager`
- `docs: update setup instructions in README`
- `test: improve coverage for RunSleep`

## 💬 Community and Questions

If you have any questions about the code architecture, how to set up your environment, or how to approach solving an Issue, don't hesitate to ask!

Our main communication channel is the **GitHub Issues**. Feel free to comment and mention the maintainer to ask for help.
