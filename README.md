# GoLang Blockchain
A blockchain implementation in Go that utilizes the UTXO (Unspent Transaction Output) model for transaction management and wallet balance computation. This project supports basic blockchain operations such as creating and printing the blockchain, managing wallets and addresses, and executing transactions. It also includes network capabilities to start nodes and mine blocks.

### Features
- Blockchain Management: Initialize and explore the blockchain with functions to create the blockchain, add blocks, and print the blockchain structure.
- Wallet Management: Generate new wallets, list existing addresses, and check balances.
- Transactions: Execute transactions between addresses, supporting the UTXO model for accurate balance tracking.
- Mining: Mine new blocks with transactions, updating the UTXO set accordingly.
- Node Operations: Start a node to participate in the network, optionally enabling mining to receive mining rewards.
- UTXO Set Management: Reindex the UTXO set for accurate and efficient balance queries.

### Environtmental Variables
NODE_ID: Specifies the ID of the node when starting a node. This is necessary for network operations and must be unique for each node in the network.
