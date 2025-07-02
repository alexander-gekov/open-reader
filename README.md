# Open Reader

A monorepo containing the Open Reader application with both frontend and backend components.

## Structure

- `apps/nuxt/` - Nuxt.js frontend application
- `apps/go/` - Go backend application

## Getting Started

### Prerequisites

- [Node.js](https://nodejs.org/) (v18 or higher)
- [pnpm](https://pnpm.io/) (v8 or higher)
- [Go](https://golang.org/) (v1.21 or higher)

### Installation

1. Install dependencies for all workspaces:
   ```bash
   pnpm install
   ```

### Development

- Start the Nuxt frontend:

  ```bash
  pnpm dev
  ```

- Or run it specifically:

  ```bash
  pnpm --filter @open-reader/nuxt dev
  ```

- Start the Go backend (from the go directory):
  ```bash
  cd apps/go
  go run .
  ```

### Building

- Build the Nuxt app:

  ```bash
  pnpm build
  ```

- Build the Go app:
  ```bash
  pnpm --filter @open-reader/go build
  ```

### Workspace Commands

- Run a command in a specific workspace:

  ```bash
  pnpm --filter @open-reader/nuxt <command>
  pnpm --filter @open-reader/go <command>
  ```

- Install a package in a specific workspace:

  ```bash
  pnpm --filter @open-reader/nuxt add <package>
  ```

- Run a command in all workspaces:
  ```bash
  pnpm -r <command>
  ```
