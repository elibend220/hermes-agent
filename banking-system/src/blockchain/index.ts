import { ethers } from 'ethers';
import config from '../config/index.js';
import logger from '../logger/index.js';

const USDT_ABI = [
  'function transfer(address to, uint256 amount) public returns (bool)',
  'function balanceOf(address account) public view returns (uint256)',
  'function approve(address spender, uint256 amount) public returns (bool)',
  'function transferFrom(address from, address to, uint256 amount) public returns (bool)',
  'event Transfer(address indexed from, address indexed to, uint256 value)',
  'event Approval(address indexed owner, address indexed spender, uint256 value)',
];

export class BlockchainService {
  private provider: ethers.JsonRpcProvider;
  private signer: ethers.Wallet;
  private usdtContract: ethers.Contract;

  constructor() {
    this.provider = new ethers.JsonRpcProvider(config.blockchain.ethereumRpcUrl);
    this.signer = new ethers.Wallet(config.blockchain.settlementPrivateKey, this.provider);
    this.usdtContract = new ethers.Contract(config.blockchain.usdtAddress, USDT_ABI, this.signer);
  }

  async transferUsdt(to: string, amount: string): Promise<string> {
    try {
      logger.info({ to, amount }, 'Initiating USDT transfer');

      const parsedAmount = ethers.parseUnits(amount, 6); // USDT has 6 decimals
      const tx = await this.usdtContract.transfer(to, parsedAmount);

      logger.info({ txHash: tx.hash }, 'USDT transfer transaction sent');
      const receipt = await tx.wait();

      logger.info({ txHash: tx.hash, blockNumber: receipt?.blockNumber }, 'USDT transfer confirmed');
      return tx.hash;
    } catch (err) {
      logger.error({ err, to, amount }, 'USDT transfer failed');
      throw err;
    }
  }

  async getBalance(address: string): Promise<string> {
    try {
      const balance = await this.usdtContract.balanceOf(address);
      return ethers.formatUnits(balance, 6);
    } catch (err) {
      logger.error({ err, address }, 'Failed to get balance');
      throw err;
    }
  }

  async getTransactionStatus(txHash: string): Promise<any> {
    try {
      const receipt = await this.provider.getTransactionReceipt(txHash);
      return {
        status: receipt ? (receipt.status === 1 ? 'confirmed' : 'failed') : 'pending',
        blockNumber: receipt?.blockNumber,
        gasUsed: receipt?.gasUsed.toString(),
        transactionHash: txHash,
      };
    } catch (err) {
      logger.error({ err, txHash }, 'Failed to get transaction status');
      throw err;
    }
  }

  async validateAddress(address: string): boolean {
    try {
      return ethers.isAddress(address);
    } catch {
      return false;
    }
  }

  async getGasPrice(): Promise<string> {
    try {
      const gasPrice = await this.provider.getFeeData();
      return ethers.formatUnits(gasPrice.gasPrice || 0, 'gwei');
    } catch (err) {
      logger.error({ err }, 'Failed to get gas price');
      throw err;
    }
  }
}

export const blockchain = new BlockchainService();
export default blockchain;
