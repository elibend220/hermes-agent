# 🤖 Artemis Personal Agent Setup Guide

**For: Eliyahu**
**Language: Hebrew/English**
**Last Updated: 2026-08-11**

---

## מה זה Artemis?

Artemis היא סוכנת AI אישית שמכירה אותך ועוזרת לך במה שאתה צריך.

הוא יודע:
- ✅ את שמך (Eliyahu)
- ✅ את ההעדפות שלך
- ✅ את משימות שאתה עושה לעתים קרובות
- ✅ ללמוד מכל אינטראקציה

---

## 🚀 התקנה מהירה

### 1️⃣ Clone את Hermes Agent

```bash
# בחר מקום לשמור את הפרויקט
cd ~/projects
git clone https://github.com/elibend220/hermes-agent.git
cd hermes-agent
```

### 2️⃣ Checkout את ה-Artemis Branch

```bash
git fetch origin artemis/jarvis-integration
git checkout artemis/jarvis-integration
# או חכה עד שזה merged ל-main:
# git pull origin main
```

### 3️⃣ התקן את Hermes

```bash
# באמצעות install script
./setup-hermes.sh

# או ידנית
uv pip install -e .
```

### 4️⃣ בדוק את הכלים של Artemis

```bash
hermes tools list | grep artemis
```

יצא כזה:
```
🔍  artemis_research
⚙️   artemis_execute
🔗  artemis_integrate
🧠  artemis_learn
📅  artemis_orchestrate
👤  artemis_profile
```

### 5️⃣ בדוק שה-Personality טוען

```bash
python3 -c "from tools.artemis_personality import ArtemisPersonality; print(ArtemisPersonality.introduction())"
```

---

## ⚙️ Personalization

### הגדרות ברירת מחדל

Artemis מגיע מוכן עבורך:
- **Name:** Eliyahu
- **Username:** eliyahu
- **Language:** Hebrew (עברית)
- **Timezone:** Asia/Jerusalem

### התאמה אישית (Optional)

צור קובץ: `~/.hermes/artemis-config.json`

```json
{
  "name": "Eliyahu",
  "username": "eliyahu",
  "language": "he",
  "timezone": "Asia/Jerusalem",
  "preferences": {
    "verbose": false,
    "use_hebrew": true,
    "emoji_style": true,
    "proactive_help": true,
    "research_depth": "moderate",
    "auto_save_learnings": true
  }
}
```

---

## 🎯 השימוש

### התחל עם Artemis

```bash
hermes
```

### פקודות בסיסיות

```bash
# ראה את הפרופיל של Artemis
> artemis_profile

# תשאל Artemis לעזרה
> אני צריך לחקור משהו בנושא X

# Artemis יצליח:
# 1. חיפוש מקורות
# 2. ניתוח עמוק
# 3. סיכום למלא

# ביקש ל-execute קוד
> הרץ script Python שעושה X

# ביקש לשלב שירות
> התחבר ל-API של Y

# ביקש להיות עזור
> תעזור לי ללמוד על Z
```

---

## 💪 Artemis Capabilities

### 🔍 Research (חיפוש ומחקר)
```bash
> אני צריך לחקור את הטרנדים האחרונים בـ AI
# Artemis יחפש, יחלץ מידע, וינתח לך
```

### ⚙️ Execute (הרצה)
```bash
> בנה לי automation script שעושה X
# Artemis יכתוב, יוודא, ויריץ קוד
```

### 🔗 Integrate (שילוב)
```bash
> התחבר ל-GitHub/Slack/Twitter שלי
# Artemis יטפל בכל ה-auth והשילוב
```

### 🧠 Learn (למידה)
```bash
> שמור את זה בזיכרון שלך
# Artemis יתעד ויזכור בשיחות הבאות
```

### 📅 Orchestrate (תיאום)
```bash
> בצע לי משימה מורכבת עם שלבים הבאים: ...
# Artemis יתכנן, יתאם, וירוץ הכל
```

---

## 🔐 Security & Privacy

Artemis:
- ✅ רץ לוקלי על המחשב שלך
- ✅ לא שולח מידע לשרת חיצוני
- ✅ שומר את הנתונים שלך בעברית
- ✅ לא אוסף מידע על משתמשים

---

## 🆘 Troubleshooting

### Artemis לא מופיע

```bash
# נקה ה-cache
rm -rf ~/.hermes/cache/tool_discovery_cache.json

# בדוק את ה-import
python3 -c "from tools.artemis_personality import ArtemisPersonality"
```

### בעיה בהתקנה

```bash
# Update Hermes
hermes update

# או reinstall
uv pip install -e . --force-reinstall
```

### צריך עזרה?

```bash
# בדוק logs
cat ~/.hermes/agent.log

# או הוסף verbose
hermes -v
```

---

## 📚 Next Steps

1. **Pull את ה-branch:** `git checkout artemis/jarvis-integration`
2. **התקן:** `./setup-hermes.sh`
3. **בדוק:** `hermes tools list | grep artemis`
4. **התחל:** `hermes`
5. **אמור:** "Artemis, הכר אותי!" 🤖

---

## 🎯 Your Artemis is Ready!

```
Welcome to Artemis, Eliyahu!

I am your personal AI agent, trained to understand your needs
and help you accomplish your goals efficiently.

How may I assist you today?
```

---

**Made with ❤️ by Claude for Eliyahu**
