# Telegram Integration

Integrate your Banking System with Telegram for real-time notifications about transactions, settlements, and system health.

## Setup Instructions

### Step 1: Create a Telegram Bot

1. Open Telegram and search for **@BotFather**
2. Send `/start` command
3. Send `/newbot` command
4. Choose a name for your bot (e.g., "Banking System Bot")
5. Choose a username for your bot (must end with "bot")
6. Copy the **API Token** (looks like: `123456789:ABCdefGHIjklmnoPQRstuvWXYZaBCDef`)

### Step 2: Get Your Chat ID

1. Open Telegram and search for **@userinfobot**
2. Send `/start` command
3. Forward any message from your banking bot to this bot
4. You'll receive your **Chat ID** (a number like: `1234567890`)

### Step 3: Configure Environment Variables

Add to your `.env` file:

```bash
TELEGRAM_BOT_TOKEN=your_bot_token_here
TELEGRAM_CHAT_ID=your_chat_id_here
```

For development, update `.env.dev`:

```bash
TELEGRAM_BOT_TOKEN=123456789:ABCdefGHIjklmnoPQRstuvWXYZaBCDef
TELEGRAM_CHAT_ID=1234567890
```

### Step 4: Restart the Server

```bash
# Stop the current server (Ctrl+C)
# Then restart:
npm run dev
# or
node dev-server.js
```

## Notifications

The system will send Telegram notifications for:

### 💳 Transaction Received
When a webhook payment is submitted:
```
💳 חדשה תרסקציה

ID: txn_abc123def456...
מ: 0x742d35Cc...
אל: 0x8ba1f109...
סכום: 1000.50 USDT
סטטוס: ✅ התקבלה

⏰ 2024-08-12 10:30:00
```

### 🏦 Settlement Batch Created
When transactions are batched for settlement:
```
🏦 יישוב אצווה

ID אצווה: batch_xyz789...
מספר תרסקציות: 5
סכום כולל: 5000.00 USDT
סטטוס: ⏳ בעיבוד

⏰ 2024-08-12 10:35:00
```

### ✅ Settlement Confirmed
When blockchain confirms the settlement:
```
✅ יישוב אושר

ID אצווה: batch_xyz789...
עסקה בבלוקצ'יין: 0xabcd1234ef567890...
סכום: 5000.00 USDT
סטטוס: ✅ אושר

👁️ צפה ב-Etherscan (clickable link)

⏰ 2024-08-12 10:38:00
```

### ⚠️ System Errors
When errors occur:
```
⚠️ שגיאה במערכת בנקאית

סוג: BLOCKCHAIN_ERROR
פרטים: Insufficient gas price
סטטוס: 🔴 דקה

⏰ 2024-08-12 10:45:00
```

### ✅ System Health
Regular health status updates:
```
✅ סטטוס מערכת בנקאית

✅ API
✅ Database
❌ Blockchain RPC
✅ Memory

⏰ 2024-08-12 11:00:00
```

## Testing Telegram Integration

Once configured, test the integration:

```bash
# Submit a test webhook
curl -X POST http://localhost:3000/api/v1/webhooks \
  -H "Content-Type: application/json" \
  -d '{
    "eventId": "evt_telegram_test",
    "data": {
      "fromAddress": "0x742d35Cc6634C0532925a3b844Bc9e7595f42472",
      "toAddress": "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
      "amount": "100",
      "currency": "USDT"
    }
  }'
```

**Expected:** You should receive a Telegram notification within seconds!

## Troubleshooting

### Bot Not Sending Messages

1. **Check credentials:**
   ```bash
   echo $TELEGRAM_BOT_TOKEN
   echo $TELEGRAM_CHAT_ID
   ```

2. **Verify bot is valid:**
   ```bash
   curl https://api.telegram.org/bot{your_token}/getMe
   ```

3. **Check server logs:**
   ```bash
   tail -f /tmp/banking-server.log | grep -i telegram
   ```

### Chat ID Issues

- Make sure Chat ID is a number (no letters)
- Forward a message from your bot to @userinfobot to get correct ID
- Don't include quotes around the Chat ID in .env

### Bot Token Issues

- Token format: `123456789:ABCdefGHIjklmnoPQRstuvWXYZaBCDef`
- Don't include brackets or extra characters
- Generate a new token if compromised: `/revoke` in @BotFather

## Advanced Usage

### Disable Telegram Notifications

Simply remove the environment variables:

```bash
# In .env
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=
```

Or comment them out:

```bash
# TELEGRAM_BOT_TOKEN=...
# TELEGRAM_CHAT_ID=...
```

### Custom Messages

You can send custom Telegram messages from your code:

```javascript
import { TelegramService } from './src/services/telegram.js';

const telegram = new TelegramService(botToken, chatId);

// Send simple message
await telegram.sendMessage('Custom alert message');

// Send formatted message
await telegram.sendMessage(`
<b>Bold Text</b>
<i>Italic Text</i>
<code>Code Block</code>
`, 'HTML');
```

### Notification Preferences

To customize which notifications you receive, edit `src/middleware/telegram-notify.ts`:

```typescript
// Comment out notifications you don't want:
// export async function notifyTransactionCreated() { ... }
// export async function notifySettlementConfirmed() { ... }
```

## Production Setup

For production, use Telegram Bot API securely:

1. **Store token in vault:** Use AWS Secrets Manager, HashiCorp Vault, etc.
2. **Use environment variables:** Never hardcode credentials
3. **Monitor message limits:** Telegram has rate limits
4. **Set up alerts:** Forward critical errors to multiple chat IDs

Example with multiple chat IDs:

```bash
# .env
TELEGRAM_BOT_TOKEN=secret_token
TELEGRAM_CHAT_ID=123456789  # Main channel
TELEGRAM_ALERTS_CHAT_ID=987654321  # Alerts only
```

## Support

For Telegram Bot questions:
- [Telegram Bot API Docs](https://core.telegram.org/bots/api)
- [BotFather Commands](https://core.telegram.org/bots#botfather)

For Banking System integration:
- Check server logs: `tail -f /tmp/banking-server.log`
- Review Telegram service: `src/services/telegram.ts`
- Check notification middleware: `src/middleware/telegram-notify.ts`
