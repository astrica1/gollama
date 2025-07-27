# Contributing to Gollama

We welcome and appreciate all contributions to the `gollama` project! Whether you're reporting a bug, suggesting a new feature, or submitting code, your help is valuable.

## How to Report a Bug

If you encounter a bug, please help us by reporting it. Before opening a new issue, please search existing issues to see if the bug has already been reported.

When creating a bug report, please include as much detail as possible:

1.  **Clear and concise description:** Explain what the bug is.
2.  **Steps to reproduce:** Provide specific steps that allow others to reliably reproduce the bug.
    * Example:
        1.  Run `go run examples/basic/main.go`
        2.  Change `Model` to "nonexistent-model"
        3.  Observe the error message.
3.  **Expected behavior:** Describe what you expected to happen.
4.  **Actual behavior:** Describe what actually happened.
5.  **Error messages/screenshots:** Include any relevant error messages from the console or screenshots.
6.  **Environment:**
    * Go version (`go version`)
    * Operating System (e.g., macOS, Windows, Linux)
    * Ollama server version (if applicable)

## How to Suggest a Feature

Have an idea for a new feature or enhancement? We'd love to hear it!

1.  **Check existing issues:** First, check if a similar feature request has already been opened.
2.  **Open a new issue:** If not, open a new issue with the label `enhancement`.
3.  **Describe the feature:**
    * Clearly explain the proposed feature.
    * Describe its benefits and how it would improve the `gollama` library.
    * Provide any relevant examples or use cases.

## Submitting a Pull Request (Code Contributions)

We appreciate code contributions that fix bugs, add new features, improve existing ones, or enhance documentation.

For **major changes** (e.g., new API endpoints, significant refactors), please **open an issue first** to discuss your ideas with the maintainers. This helps ensure your effort aligns with the project's direction and avoids duplicated work.

For **minor changes** (e.g., typo fixes, small bug fixes, minor documentation updates), feel free to submit a pull request directly.

Here's the general process for submitting a pull request:

1.  **Fork the repository:** Click the "Fork" button on the top right of the [gollama GitHub repository](https://github.com/astrica1/gollama). This creates a copy of the repository in your GitHub account.
2.  **Clone your forked repository:**
    ```bash
    git clone [https://github.com/YOUR_USERNAME/gollama.git](https://github.com/YOUR_USERNAME/gollama.git)
    cd gollama
    ```
    (Replace `YOUR_USERNAME` with your GitHub username.)
3.  **Create a new branch:** Always create a new branch for your changes. Use a descriptive name (e.g., `feat/add-new-feature` or `bugfix/fix-issue-123`).
    ```bash
    git checkout -b your-new-branch-name
    ```
4.  **Make your changes:** Implement your bug fix, feature, or documentation update.
    * Ensure your code follows Go best practices and the existing coding style.
    * Add or update tests for new functionality or bug fixes to maintain comprehensive test coverage.
5.  **Run tests and linters:**
    ```bash
    go test ./...
    go fmt ./...
    golangci-lint run # If golangci-lint is installed.
    ```
    Ensure all tests pass and your code is formatted correctly.
6.  **Commit your changes:** Write clear, concise, and descriptive commit messages. A good commit message explains *what* changed and *why*.
    ```bash
    git add .
    git commit -m "feat: Briefly describe your feature or fix"
    ```
7.  **Push your branch to your fork:**
    ```bash
    git push origin your-new-branch-name
    ```
8.  **Open a Pull Request (PR):**
    * Go to your forked repository on GitHub.
    * You should see a banner or button prompting you to "Compare & pull request" for your newly pushed branch. Click it.
    * Ensure the base repository is `astrica1/gollama` and the base branch is `main`.
    * Provide a clear and concise title for your PR.
    * In the description, explain your changes in detail.
    * **Reference any related issues** by including `#ISSUE_NUMBER` (e.g., `Fixes #123`, `Closes #456`, `Addresses #6`). This will link your PR to the issue.
    * Submit your pull request.

## Development Setup

To get your development environment ready:

* **Requirements:**
    * Go 1.21 or later
    * Access to an Ollama server (local or remote)
* **Running tests:**
    ```bash
    go test ./...
    ```
* **Running examples:**
    ```bash
    go run examples/basic/main.go
    # Check other examples in the examples/ directory
    ```
* **Formatting code:**
    ```bash
    go fmt ./...
    ```
* **Linting code:** (Requires `golangci-lint` to be installed: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)
    ```bash
    golangci-lint run
    ```

Thank you for contributing to Gollama!