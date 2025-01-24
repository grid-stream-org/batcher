# <p align="center">Batcher - Micro-batching Data Processing Service</p>

<p align="center"><img src="assets/logo.svg" width="350px"/></p>
<p align="center">The micro-batching data-processing Go service responsible for buffering incoming data from a MQTT broker, normalizing it, and writing to a destination.</p>

## 🧭 Table of Contents

- [Batcher - Micro-batching Data-processing Service](#batcher---micro-batching-data-processing-service)
  - [Table of Contents](#-table-of-contents)
  - [Team](#-team)
  - [Directory Structure](#-directory-structure)
  - [Contributing](#-contributing)
  - [Local Run](#-local-run)
    - [Prerequisites](#prerequisites)
    - [Steps](#steps)

## 👥 Team

| Team Member     | Role Title                | Description                                                                                                                                             |
| --------------- | ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Matthew Collett | Technical Lead/Developer  | Focus on architecture design and solving complex problems, with a focus on the micro-batching process.                                                  |
| Cooper Dickson  | Project Manager/Developer | Ensure that the scope and timeline are feasible and overview project status, focus on UI and real-time transmission.                                    |
| Eric Cuenat     | Scrum Master/Developer    | In charge of agile methods for the team such as organizing meetings, removing blockers, and team communication, focus on UI and web socket interaction. |
| Sam Keays       | Product Owner/Developer   | Manager of product backlog and updating board to reflect scope changes and requirements, focus on database operations and schema design.                |

## 🏗️ Directory Structure

- `.github/`
  - GitHub CI/CD workflows
- `assets/`
  - Static assets
- `cmd/`
  - Main application
- `configs/`
  - Configuration file template
- `internal/`
  - Private application and library code
- `metrics/`
  - Global Prometheus metrics instance
- `pkg/`
  - Public application and library code

## ⛑️ Contributing

For guidlines and instructions on contributing, please refer to [CONTRIBUTING.md](https://github.com/grid-stream-org/batcher/blob/main/CONTRIBUTING.md)

## 🚀 Local Run

### Prerequisites
- Ensure you have go installed

### Steps
1. First, start by cloning this repository to your local machine
```bash
git clone https://github.com/grid-stream-org/batcher.git
```
2. Navigate to the project directory
```bash
cd batcher
```
3. Populate `configs/config.json` with valid config to run the batcher
4. Install the project dependencies
```bash
make download
```
5. Run the batcher
```bash
make run
```
