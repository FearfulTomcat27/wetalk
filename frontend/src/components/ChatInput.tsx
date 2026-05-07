"use client";

import { useState, useRef, useCallback, type KeyboardEvent } from "react";

interface ChatInputProps {
  onSend: (content: string) => void;
  disabled?: boolean;
}

export function ChatInput({ onSend, disabled = false }: ChatInputProps) {
  const [text, setText] = useState("");
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const handleSend = useCallback(() => {
    if (disabled || text.trim() === "") {return;}
    onSend(text.trim());
    setText("");
    // 重置高度
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
    }
  }, [text, onSend, disabled]);

  const handleKeyDown = useCallback(
    (e: KeyboardEvent<HTMLTextAreaElement>) => {
      // Enter 发送，Shift+Enter 换行
      if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault();
        handleSend();
      }
    },
    [handleSend],
  );

  // 自动调整高度
  const handleInput = useCallback(() => {
    const el = textareaRef.current;
    if (!el) {return;}
    el.style.height = "auto";
    el.style.height = `${Math.min(el.scrollHeight, 120)}px`;
  }, []);

  const canSend = text.trim() !== "" && !disabled;

  return (
    <div className="shrink-0 border-t bg-card px-4 py-3">
      <div className="mx-auto flex max-w-3xl items-end gap-3">
        {/* 输入区域 */}
        <div className="flex flex-1 items-end rounded-2xl border bg-muted/40 px-4 py-2.5 transition-colors focus-within:border-primary/30 focus-within:bg-background focus-within:ring-2 focus-within:ring-primary/10">
          <textarea
            ref={textareaRef}
            value={text}
            onChange={(e) => {
              setText(e.target.value);
              handleInput();
            }}
            onKeyDown={handleKeyDown}
            placeholder={
              disabled ? "请先选择联系人" : "输入消息… (Enter 发送，Shift+Enter 换行)"
            }
            disabled={disabled}
            rows={1}
            className="max-h-[120px] min-h-[24px] w-full resize-none bg-transparent text-sm text-foreground placeholder:text-muted-foreground/50 focus:outline-none disabled:cursor-not-allowed disabled:opacity-50"
          />
        </div>

        {/* 发送按钮 - 圆形带图标 */}
        <button
          type="button"
          onClick={handleSend}
          disabled={!canSend}
          className="flex size-10 shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground shadow-sm transition-all hover:bg-primary/90 active:scale-95 disabled:cursor-not-allowed disabled:opacity-40 disabled:active:scale-100"
          aria-label="发送消息"
        >
          <svg
            className="size-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
            />
          </svg>
        </button>
      </div>
    </div>
  );
}
