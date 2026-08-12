import logger from '../logger/index.js';
import config from '../config/index.js';

export interface TelegramConfig {
  botToken: string;
  chatId: string;
}

export class TelegramService {
  private botToken: string;
  private chatId: string;
  private apiUrl = 'https://api.telegram.org';

  constructor(botToken: string, chatId: string) {
    this.botToken = botToken;
    this.chatId = chatId;
  }

  async sendMessage(text: string, parseMode: 'HTML' | 'Markdown' = 'HTML'): Promise<boolean> {
    try {
      const url = `${this.apiUrl}/bot${this.botToken}/sendMessage`;

      const response = await fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          chat_id: this.chatId,
          text,
          parse_mode: parseMode,
          disable_web_page_preview: true,
        }),
      });

      if (!response.ok) {
        logger.error({ status: response.status }, 'Telegram message failed');
        return false;
      }

      logger.info({ text: text.substring(0, 50) }, 'Telegram message sent');
      return true;
    } catch (err) {
      logger.error({ err }, 'Telegram service error');
      return false;
    }
  }

  async notifyTransaction(txId: string, fromAddr: string, toAddr: string, amount: string, currency: string) {
    const message = `
<b>💳 חדשה תרסקציה</b>

<b>ID:</b> <code>${txId}</code>
<b>מ:</b> <code>${fromAddr.substring(0, 10)}...</code>
<b>אל:</b> <code>${toAddr.substring(0, 10)}...</code>
<b>סכום:</b> ${amount} ${currency}
<b>סטטוס:</b> ✅ התקבלה

⏰ <i>${new Date().toLocaleString('he-IL')}</i>
    `.trim();

    return this.sendMessage(message);
  }

  async notifySettlement(batchId: string, transactionCount: number, totalAmount: string) {
    const message = `
<b>🏦 יישוב אצווה</b>

<b>ID אצווה:</b> <code>${batchId}</code>
<b>מספר תרסקציות:</b> ${transactionCount}
<b>סכום כולל:</b> ${totalAmount} USDT
<b>סטטוס:</b> ⏳ בעיבוד

⏰ <i>${new Date().toLocaleString('he-IL')}</i>
    `.trim();

    return this.sendMessage(message);
  }

  async notifySettlementConfirmed(batchId: string, txHash: string, totalAmount: string) {
    const message = `
<b>✅ יישוב אושר</b>

<b>ID אצווה:</b> <code>${batchId}</code>
<b>עסקה בבלוקצ'יין:</b> <code>${txHash}</code>
<b>סכום:</b> ${totalAmount} USDT
<b>סטטוס:</b> ✅ אושר

<a href="https://etherscan.io/tx/${txHash}">👁️ צפה ב-Etherscan</a>

⏰ <i>${new Date().toLocaleString('he-IL')}</i>
    `.trim();

    return this.sendMessage(message);
  }

  async notifyError(errorType: string, details: string) {
    const message = `
<b>⚠️ שגיאה במערכת בנקאית</b>

<b>סוג:</b> ${errorType}
<b>פרטים:</b> <code>${details}</code>
<b>סטטוס:</b> 🔴 דקה

⏰ <i>${new Date().toLocaleString('he-IL')}</i>
    `.trim();

    return this.sendMessage(message);
  }

  async notifyHealth(status: 'healthy' | 'degraded' | 'unhealthy', checks: Record<string, boolean>) {
    const statusEmoji = {
      healthy: '✅',
      degraded: '⚠️',
      unhealthy: '🔴',
    };

    const checksList = Object.entries(checks)
      .map(([name, ok]) => `${ok ? '✅' : '❌'} ${name}`)
      .join('\n');

    const message = `
<b>${statusEmoji[status]} סטטוס מערכת בנקאית</b>

${checksList}

⏰ <i>${new Date().toLocaleString('he-IL')}</i>
    `.trim();

    return this.sendMessage(message);
  }

  async notifyWebhookReceived(eventId: string, amount: string) {
    const message = `
<b>📲 ווהוק התקבל</b>

<b>ID:</b> <code>${eventId}</code>
<b>סכום:</b> ${amount} USDT
<b>סטטוס:</b> ✅ בעיבוד

⏰ <i>${new Date().toLocaleString('he-IL')}</i>
    `.trim();

    return this.sendMessage(message);
  }
}

export const createTelegramService = (botToken?: string, chatId?: string) => {
  const token = botToken || process.env.TELEGRAM_BOT_TOKEN;
  const chat = chatId || process.env.TELEGRAM_CHAT_ID;

  if (!token || !chat) {
    logger.warn('Telegram service not configured (missing TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID)');
    return null;
  }

  return new TelegramService(token, chat);
};

export default TelegramService;
