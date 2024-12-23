### Project Overview

**Forwarder** is a transaction forwarding service, primarily designed for Solana, where it manages payment creation, forwarding transactions, and webhook callbacks. It listens for incoming requests to create payment addresses and amounts, then forwards the transactions to a configured primary address, while also notifying a webhook with transaction details.

---

### Features
- **Transaction Creation**: Endpoint `/payment/create` generates a new payment address and amount to send.
- **Webhook Notifications**: Sends transaction details to a configured webhook once the payment is made.
- **Flexible Configurations**: Environment-based configurations for ease of deployment across various environments.
- **Scalable**: Modular architecture for easy extension and maintainability.
- **Secure Handling**: Ensures secure management of transaction details.

---

### Data Structures

Here are the key data structures used in the application:

#### `PaymentCreateBody`
This is the body used when creating a new payment request.

- **amount**: The amount to send in the transaction.
- **callback_uri**: The URL to notify with transaction details once the payment is processed.

#### `PaymentCreateResponse`
This is the response returned when a payment address is created.

- **success**: Indicates if the payment creation was successful.
- **id**: Unique identifier for the transaction.
- **amount**: The amount to be sent in the transaction.
- **address**: The address generated to which the payment should be sent.
- **qrcode**: QR code data in base64 for the generated address (optional but useful for mobile wallets).
- **expires**: Timestamp for when the payment link expires.

#### `WebhookResponse`
This is the structure used to send transaction details to a webhook after a successful payment.

- **success**: Indicates if the transaction was successfully completed.
- **id**: The unique identifier of the payment request.
- **error**: Any error that occurred during the transaction.
- **desired_amount**: The amount that was requested to be sent.
- **amount_sent**: The actual amount that was sent.
- **transaction_id**: The unique ID for the transaction.
- **address**: The payment address to which the transaction was sent.
- **time_sent**: The timestamp when the transaction was sent.
- **percent_of_total**: Percentage of the total expected amount that was sent.

---

### Prerequisites

- Go 1.18 or later.
- MongoDB via localhost or cloud (Optional)
- Proper configuration via environment variables.

### Installation

1. **Clone the repository**:

```bash
git clone https://github.com/Aran404/Forwarder.git
cd Forwarder
```

2. **Install dependencies**:

```bash
go mod tidy
```

3. **Set up environment variables**:

- Copy `.env.sample` to `.env`:

```bash
cp .env.sample .env
```

- Edit the `.env` file to set solana cluster.

4. **Build the project**:

```bash
go build -o api/cmd .
```

### Usage

To start the application, either run the compiled binary or run the main.go in api/cmd

```bash
./bin/forwarder
```

This will start the application, and it will listen for incoming requests on the default address (`localhost:3443`).

- To create a payment, send a `POST` request to `localhost:3443/payment/create` with the `Amount` and `CallbackURI` as a JSON body.
  
  Example request:
  ```json
  {
    "amount": 10.0, // In Solana
    "callback_uri": "http://example.com/webhook"
  }
  ```

- The response will provide the payment address, amount, and a QR code to complete the transaction.
- Once the payment is sent, the application will notify the configured `CallbackURI` with transaction details.

### Contributing

Contributions are welcome! Please fork the repository, create a new branch, make your changes, and submit a pull request.

### License

This project is licensed under the GPL 3.0 License. See the [LICENSE](https://choosealicense.com/licenses/gpl-3.0/) file for details.