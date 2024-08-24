# Terraform Provider for QNAP

This repository contains the Terraform provider for managing QNAP Container Station containers and apps. It enables users to interact with QNAP's Container Station APIs to manage and automate configurations on QNAP Container Station containers through Terraform.

## Features

- **Manage QNAP Resources**: Create, update, and delete QNAP Container Station resources like, containers and applications
- **Data Sources**: Retrieve information about existing QNAP Container Station resources.
- **Automation**: Automate the management of your QNAP NAS infrastructure using Terraform's declarative configuration language.

## Requirements

- **Terraform**: `v0.12+`
- **Go**: `v1.17+` (for building the provider)

## Installation

### Using the Terraform Registry (Recommended)

This provider is available on the Terraform Registry. To use it, add the following block to your Terraform configuration:

```hcl
terraform {
  required_providers {
    qnap = {
      source = "mohamed-mfarag/qnap"
      version = "0.5.0"
    }
  }
}

provider "qnap" {
  # Configuration options
}
```

### Building the Provider Locally

If you prefer to build the provider from source, follow these steps:

1. **Clone the repository**:

   ```sh
   git clone https://github.com/mohamed-mfarag/terraform-provider-qnap.git
   cd terraform-provider-qnap
   ```

2. **Build the provider**:

   ```sh
   go build -o terraform-provider-qnap
   ```

3. **Move the binary to your Terraform plugins directory**:

   ```sh
   mkdir -p ~/.terraform.d/plugins/mohamed-mfarag/qnap/<tag>/linux_amd64
   mv terraform-provider-qnap ~/.terraform.d/plugins/mohamed-mfarag/qnap/0.5.0/linux_amd64/
   ```

## Usage

### Example Configuration

Below is a simple example of how to use the QNAP provider to manage a QNAP Container Station container:

```hcl
provider "qnap" {
  hostname = "https://qnap.example.com"
  username = "admin"
  password = "your-password"
}

resource "qnap_container" "ubuntu-1" {
  name              = "ubuntu-10"
  image             = "ubuntu:latest"
  type              = "docker"
  network           = "bridge"
  networktype       = "default"
  removeanonvolumes = true
}

```

More information and samples under the [docs](docs) section

### Running Terraform

Once your configuration is ready, you can initialize and apply the Terraform configuration:

```sh
terraform init
terraform apply
```

## Running Acceptance Tests

If you are contributing to the provider and want to run the acceptance tests:

1. **Set up your environment**: Ensure that your QNAP device is accessible and that the necessary environment variables are configured:

   ```sh
   export QNAP_HOSTNAME="your-qnap-device"
   export QNAP_USERNAME="admin"
   export QNAP_PASSWORD="your-password"
   ```

2. **Run the tests**:

   ```sh
   TF_CLI_ARGS_apply="-parallelism=1" TF_ACC=1 go test -count=1 -v
   ```

   Note: Running acceptance tests on CI platforms like GitHub Actions might require network access to the QNAP device, which could be challenging. It's recommended to mock the tests or run them in a self-hosted environment.

## Contributing

Contributions to this project are welcome! To contribute:

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/my-feature`).
3. Commit your changes (`git commit -am 'Add new feature'`).
4. Push to the branch (`git push origin feature/my-feature`).
5. Open a pull request.

## License

This project is licensed under the Apache License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Special thanks to the open-source community and contributors.