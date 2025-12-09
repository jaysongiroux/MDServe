<!-- 
{
    "tags": ["example"]
}
-->
# 📘 Introduction to Markdown

*A lightweight markup language for humans*

Markdown is a simple, readable way to format text that converts cleanly into HTML.
It’s popular for documentation, README files, blogs, and note-taking systems.

In this short guide, you’ll learn:

* What Markdown is
* How to write common formatting
* Code blocks, lists, images, quotes
* Tips for writing clean Markdown

---

## 🧩 What Is Markdown?

Markdown is a *markup language* created by John Gruber in 2004.
Its purpose: **write text that looks good as plain text, and even better when rendered.**

For example:

```
This is *italic*, this is **bold**, and this is `inline code`.
```

---

## ✍️ Basic Formatting

### **Bold & Italic**

```md
*italic*
**bold**
***bold italic***
```

### **Headings**

```md
# H1
## H2
### H3
###### H6
```

### **Links**

```md
[OpenAI](https://openai.com)
```

### **Images**

```md
![Alt text](https://example.com/image.png)
```

---

## 📜 Lists

### **Unordered**

```md
- Item one
- Item two
  - Nested item
```

### **Ordered**

```md
1. First
2. Second
3. Third
```

---

## 💬 Blockquotes

```md
> Markdown is easy to learn.
>  
> — Someone on the internet
```

---

## 🧱 Code Blocks

You can add multiline code using triple backticks:

<pre>
```go
func main() {
    fmt.Println("Hello, Markdown!")
}
```
</pre>

---

## 📑 Tables

```md
| Language | Typed | Popularity |
|----------|-------|------------|
| Go       | Yes   | High       |
| JavaScript | No  | Huge       |
```

---

## Mermaid Diagrams

```mermaid
graph TD;
    A-->B;
    A-->C;
    B-->D;
    C-->D;
```

---

## 🧠 Tips for Clean Markdown

* Leave one blank line between sections
* Use consistent heading levels
* Prefer lists over long paragraphs
* Use code blocks for anything technical
* Keep line length readable (80–120 chars)

---

## ✅ Conclusion

Markdown is extremely easy to learn but powerful enough for documentation, articles, and entire websites. You write your content once—tools convert it anywhere.

Happy writing!
