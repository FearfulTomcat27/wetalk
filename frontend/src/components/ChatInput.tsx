"use client";

import {
  useState,
  useRef,
  useCallback,
  forwardRef,
  useImperativeHandle,
  type KeyboardEvent,
} from "react";
import { Smile, Paperclip, Image } from "lucide-react";

interface ChatInputProps {
  onSend: (content: string) => void;
  disabled?: boolean;
  /** 受控值 — 父组件管理（切换联系人自动切换输入文本） */
  value?: string;
  /** 受控 onChange */
  onChange?: (text: string) => void;
}

const ChatInput = forwardRef<HTMLTextAreaElement, ChatInputProps>(
  function ChatInput({ onSend, disabled = false, value, onChange }, ref) {
    const [localText, setLocalText] = useState("");
    const innerRef = useRef<HTMLTextAreaElement>(null);

    // 受控 / 非受控
    const isControlled = value !== undefined;
    const text = isControlled ? value : localText;

    // 将内部 textarea ref 暴露给父组件
    useImperativeHandle<HTMLTextAreaElement | null, HTMLTextAreaElement | null>(
      ref,
      () => innerRef.current,
    );

    const handleInput = useCallback(() => {
      const el = innerRef.current;
      if (!el) {return;}
      el.style.height = "auto";
      el.style.height = `${Math.min(el.scrollHeight, 120)}px`;
    }, []);

    function handleChange(e: React.ChangeEvent<HTMLTextAreaElement>) {
      const newText = e.target.value;
      if (isControlled) {
        onChange?.(newText);
      } else {
        setLocalText(newText);
      }
      handleInput();
    }

    const handleSend = useCallback(() => {
      if (disabled || text.trim() === "") {return;}
      onSend(text.trim());
      if (!isControlled) {
        setLocalText("");
      }
      if (innerRef.current) {
        innerRef.current.style.height = "auto";
      }
    }, [text, onSend, disabled, isControlled]);

    const handleKeyDown = useCallback(
      (e: KeyboardEvent<HTMLTextAreaElement>) => {
        if (e.key === "Enter" && !e.shiftKey) {
          e.preventDefault();
          handleSend();
        }
      },
      [handleSend],
    );

    const toolbarButtons = [
      { icon: Smile, label: "表情" },
      { icon: Paperclip, label: "文件" },
      { icon: Image, label: "图片" },
    ];

    return (
      <div className="shrink-0 px-4 py-3">
        <div className="mx-auto">
          <div className="overflow-hidden rounded-2xl bg-muted/40 ring-1 ring-border/50 transition-shadow focus-within:ring-2 focus-within:ring-primary/20">
            <textarea
              ref={innerRef}
              value={text}
              onChange={handleChange}
              onKeyDown={handleKeyDown}
              placeholder={
                disabled ? "请先选择联系人" : "输入消息… (Enter 发送，Shift+Enter 换行)"
              }
              disabled={disabled}
              rows={2}
              className="block w-full resize-none bg-transparent px-4 pt-3 text-sm text-foreground placeholder:text-muted-foreground/50 focus:outline-none disabled:cursor-not-allowed disabled:opacity-50"
              style={{ minHeight: "4rem", maxHeight: "120px" }}
            />

            <div className="flex items-center gap-4 px-4 pb-3 pt-1">
              {toolbarButtons.map(({ icon: Icon, label }) => (
                <button
                  key={label}
                  type="button"
                  title={label}
                  className="flex size-8 items-center justify-center rounded-full text-muted-foreground/50 transition-colors hover:bg-muted-foreground/10 hover:text-muted-foreground"
                >
                  <Icon className="size-[18px]" />
                </button>
              ))}
            </div>
          </div>
        </div>
      </div>
    );
  },
);

export { ChatInput };
