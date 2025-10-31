## BishopFox MCP Prototype

This is a prototype to demonstrate tool implementation via MCP and Bedrock RETURN_CONTROL.

config/setup.sql contains the database schema including test data from generate_fixtures.py.

## Running the Prototype in a container

Create `.app.env` according to `.app.env.example`. It needs to be configured with AWS
secrets to access the Bedrock API.

Run `docker compose up` to start the app.

The MCP server is hosted at `http://localhost:8112/mcp` using Streamable HTTP transport.

The HTTP server is hosted at `http://localhost:8110/`.

HTTP endpoints:
- `POST /ask?organization_id=<orgid>` 
  - Body: {query: "question to ask"}

## Additional notes

If running outside of the container, use the `.env` file. VSCode configurations can be set
to load the `.env` file.

.vscode/launch.json:
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Program",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}",
            "envFile": "${workspaceFolder}/.env"
        }
    ]
}
```

.vscode/settings.json:
```json
{
    "go.testEnvFile": "${workspaceFolder}/.env",
    "go.delveConfig": {
        "envFile": "${workspaceFolder}/.env"
    }
}
```
