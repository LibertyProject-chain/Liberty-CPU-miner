# Liberty Project Miner

Welcome to **Liberty Project Miner**! This repository contains the Liberty Project network CPU miner. The network coin is **LBRT** and this miner allows members to contribute to the network by mining blocks and receiving rewards.

Note: the external miner finds solutions and reports them to the node. if the node accepts the solution, the reward is credited to the node's coinbase.

---



## Downloads

**Latest Release v0.6.8:**

- [Linux (amd64)](https://github.com/LibertyProject-chain/Liberty-CPU-miner/releases/download/v0.6.7/liberty-linux-amd64)
- [Windows (amd64)](https://github.com/LibertyProject-chain/Liberty-CPU-miner/releases/download/v0.6.7/liberty-windows-amd64.exe)

---

## How to Start Mining

### Basic Command

```bash
./miner <url-rpc> <number-of-threads>
```

### Example

```bash
./liberty-linux-amd64 http://ip_node:9945 12
```
or

```bash
liberty-windows-amd64.exe http://ip_node:9945 12
```

### Parameters

- **`<url-rpc>`**: The RPC URL of your node to which you want to connect.
- **`<number-of-threads>`**: The number of CPU threads to use for mining.
 
Note: it is important for a miner to use its own rpc configured by [Liberty Project chain](https://github.com/LibertyProject-chain/LibertyProject-chain)

---

## Requirements

- **Hardware**: x86-64 CPU
- **Operating System**: Linux or Windows (amd64)
- **Dependencies**: Ensure that your environment supports the provided binaries.

---

## Contribution

This miner is part of the Liberty Project. We welcome feedback and contributions to improve the miner and the ecosystem.

Feel free to create issues or submit pull requests for enhancements.

---

Happy mining! 🚀

