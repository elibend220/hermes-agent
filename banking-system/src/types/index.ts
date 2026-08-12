export interface Transaction {
  id: string;
  eventId: string;
  fromAddress: string;
  toAddress: string;
  amount: string;
  currency: string;
  status: 'pending' | 'confirmed' | 'failed' | 'received';
  blockchainTxHash?: string;
  createdAt: Date;
  updatedAt: Date;
  metadata?: Record<string, any>;
}

export interface WebhookPayload {
  eventId: string;
  timestamp: string;
  type: 'payment' | 'settlement' | 'reconciliation';
  data: {
    fromAddress: string;
    toAddress: string;
    amount: string;
    currency: 'USDT' | 'USD';
    reference?: string;
    description?: string;
  };
}

export interface SettlementBatch {
  id: string;
  batchNumber: number;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  transactionCount: number;
  totalAmount: string;
  createdAt: Date;
  completedAt?: Date;
  blockchainTxHash?: string;
}

export interface ApiKeyRequest {
  headers: {
    'x-api-key': string;
    'content-type': string;
  };
}

export interface User {
  id: string;
  email: string;
  walletAddress: string;
  apiKey: string;
  isActive: boolean;
  createdAt: Date;
  updatedAt: Date;
}

export interface ErrorResponse {
  success: false;
  error: {
    code: string;
    message: string;
    statusCode: number;
  };
}

export interface SuccessResponse<T> {
  success: true;
  data: T;
  timestamp: string;
}
