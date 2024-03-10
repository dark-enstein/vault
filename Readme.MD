# Vault - Your Mini Secret Manager

Welcome to **Vault**, the compact and versatile secret manager designed to securely manage your sensitive data such as secrets, tokens, and other critical information. Whether you're a developer needing to safeguard API keys or a system administrator managing access credentials, Vault has got you covered .

Access the GoDoc: **[Documentation](https://pkg.go.dev/github.com/dark-enstein/vault)**

## Key Features and Functionalities

- **Secure Storage**: Encrypt and store sensitive information securely using various storage backends including local file systems, Redis servers, and in-memory storage.
- **Flexible Retrieval**: Easily retrieve and decrypt your stored secrets with straightforward commands.
- **CLI & Service Operation**: Use Vault as a handy CLI tool for direct interaction or as a resilient background service.
- **Multi-Storage Backend Support**: Initialize Vault with your choice of storage backend, catering to your specific needs and environment.
- **Easy Management**: List, store, delete, peek, and peel tokens with simple commands.
- **Debug Mode**: Toggle debug mode for detailed operational logging.

## Operation

Vault can be utilized in two main ways: as a CLI tool for direct interaction or as a service for continuous, background operation.

### As a CLI Tool

Vault offers an array of commands for managing secrets directly from your terminal. Here are some examples to get you started:

- Initialize Vault: `vault init --store redis --connectionString "redis://user:password@localhost:6379"`
- Store a new secret: `vault store --id "myTokenID" --secret "sensitiveValue"`
- Retrieve a decrypted token: `vault peel --id "myTokenID"`
- Delete a token: `vault delete --id "myTokenID"`
- List all tokens: `vault list`

### As a Service

Run Vault in the background as a service to manage secrets seamlessly across your infrastructure.

- Start the service: `vault service run`

## Installation

### Using Makefile

- **Simple Install**: Run `make` to compile and build Vault.
- **Installing with Auto Configuration**: Use `make sudo` for automatic installation and configuration, including adding Vault to your `$PATH` and cleaning up post-build.

## Usage

After installation, Vault is ready to safeguard your secrets. Whether you prefer the command line for quick interactions or setting up Vault as a persistent service, managing your sensitive data has never been easier.

## Contributing

Contributions are welcome! If you're looking to contribute to Vault, please follow our guidelines for code style, formatting, and testing. Fork the repository, make your changes, and submit a pull request. We appreciate your input in making Vault even better!

## License

Vault is distributed under [MIT License](https://opensource.org/licenses/MIT). Feel free to use, modify, and distribute as per the license terms.

## Contact

For questions, suggestions, or contributions, please contact the project maintainers at [vault@greystein.com](mailto:vault@greystein.com). Your feedback is valuable to us!

## Additional Information

For more details, documentation, or to view the source code, visit [Vault on GitHub](https://github.com/dark-enstein/vault). Check out related projects and join our community forum for support and discussions.

Vault has evolved from [ls](https://github.com/dark-enstein/ls) so its independent.
