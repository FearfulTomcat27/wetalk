"use client";

import {
  useState,
  useRef,
  useCallback,
  forwardRef,
  useImperativeHandle,
  useEffect,
  type KeyboardEvent,
} from "react";
import { Smile, Paperclip, ImageIcon, Loader2 } from "lucide-react";
import { uploadFile } from "@/lib/api/upload";

/** 常用 emoji 列表 */
const EMOJI_LIST = [
  "😀", "😃", "😄", "😁", "😆", "😅", "🤣", "😂", "🙂", "😊",
  "😇", "😍", "🥰", "😘", "😗", "😋", "😛", "😜", "🤪", "😝",
  "🤑", "🤗", "🤭", "🤫", "🤔", "🤐", "🤨", "😐", "😑", "😶",
  "😏", "😒", "🙄", "😬", "😮", "😯", "😲", "😳", "🥺", "😦",
  "😧", "😨", "😰", "😥", "😢", "😭", "😱", "😖", "😣", "😡",
  "😠", "🤯", "💪", "👍", "👎", "👏", "🙏", "🤝", "❤️", "💔",
  "💯", "🔥", "⭐", "🎉", "🎊", "💐", "🌹", "☕", "🎵", "🎶",
];

const MAX_IMAGE_SIZE = 10 * 1024 * 1024; // 10MB
const MAX_FILE_SIZE = 20 * 1024 * 1024; // 20MB

interface ChatInputProps {
  onSend: (content: string) => void;
  /** 发送富媒体消息（图片/文件） */
  onSendMedia?: (content: string, contentType: string, fileMetadata?: { url: string; original_name: string; file_size: number; mime_type: string }) => void;
  disabled?: boolean;
  /** 受控值 — 父组件管理（切换联系人自动切换输入文本） */
  value?: string;
  /** 受控 onChange */
  onChange?: (text: string) => void;
}

const ChatInput = forwardRef<HTMLTextAreaElement, ChatInputProps>(
  function ChatInput({ onSend, onSendMedia, disabled = false, value, onChange }, ref) {
    const [localText, setLocalText] = useState("");
    const [emojiOpen, setEmojiOpen] = useState(false);
    const [uploading, setUploading] = useState(false);
    const innerRef = useRef<HTMLTextAreaElement>(null);
    const emojiPanelRef = useRef<HTMLDivElement>(null);
    const imageInputRef = useRef<HTMLInputElement>(null);
    const fileInputRef = useRef<HTMLInputElement>(null);

    // 受控 / 非受控
    const isControlled = value !== undefined;
    const text = isControlled ? value : localText;

    // 将内部 textarea ref 暴露给父组件
    useImperativeHandle<HTMLTextAreaElement | null, HTMLTextAreaElement | null>(
      ref,
      () => innerRef.current,
    );

    // 点击外部关闭 emoji 面板
    useEffect(() => {
      if (!emojiOpen) {return;}
      function handleClickOutside(e: MouseEvent) {
        if (
          emojiPanelRef.current &&
          !emojiPanelRef.current.contains(e.target as Node)
        ) {
          setEmojiOpen(false);
        }
      }
      document.addEventListener("mousedown", handleClickOutside);
      return () => document.removeEventListener("mousedown", handleClickOutside);
    }, [emojiOpen]);

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

    function handleEmojiSelect(emoji: string) {
      const newText = text + emoji;
      if (isControlled) {
        onChange?.(newText);
      } else {
        setLocalText(newText);
      }
      setEmojiOpen(false);
      innerRef.current?.focus();
    }

    async function handleFileSelect(files: FileList | null, type: "image" | "file") {
      if (!files || files.length === 0) {return;}
      const file = files[0];
      const maxSize = type === "image" ? MAX_IMAGE_SIZE : MAX_FILE_SIZE;
      if (file.size > maxSize) {
        const label = type === "image" ? "图片" : "文件";
        const limit = type === "image" ? "10MB" : "20MB";
        alert(`${label}大小不能超过${limit}`);
        return;
      }
      setUploading(true);
      try {
        const res = await uploadFile(file, type);
        const url = res.data!.url;
        const fileMetadata = {
          url,
          original_name: file.name,
          file_size: file.size,
          mime_type: file.type || (type === "image" ? "image/png" : "application/octet-stream"),
        };
        onSendMedia?.(url, type, fileMetadata);
      } catch {
        // 错误已在拦截器 toast
      } finally {
        setUploading(false);
        // 清空 input value 以便重复选择同一文件
        if (type === "image" && imageInputRef.current) {
          imageInputRef.current.value = "";
        } else if (fileInputRef.current) {
          fileInputRef.current.value = "";
        }
      }
    }

    const handleSend = useCallback(() => {
      if (disabled || uploading || text.trim() === "") {return;}
      onSend(text.trim());
      if (!isControlled) {
        setLocalText("");
      }
      if (innerRef.current) {
        innerRef.current.style.height = "auto";
      }
    }, [text, onSend, disabled, uploading, isControlled]);

    const handleKeyDown = useCallback(
      (e: KeyboardEvent<HTMLTextAreaElement>) => {
        if (e.key === "Enter" && !e.shiftKey) {
          e.preventDefault();
          handleSend();
        }
      },
      [handleSend],
    );

    return (
      <div className="shrink-0 px-4 py-3">
        {/* 隐藏文件选择器 */}
        <input
          ref={imageInputRef}
          type="file"
          accept="image/*"
          className="hidden"
          onChange={(e) => handleFileSelect(e.target.files, "image")}
        />
        <input
          ref={fileInputRef}
          type="file"
          className="hidden"
          onChange={(e) => handleFileSelect(e.target.files, "file")}
        />

        <div className="mx-auto relative">
          {/* Emoji 面板 */}
          {emojiOpen && (
            <div
              ref={emojiPanelRef}
              className="absolute bottom-full left-0 mb-2 rounded-xl bg-card p-3 shadow-lg ring-1 ring-border/50"
            >
              <div className="grid grid-cols-10 gap-1">
                {EMOJI_LIST.map((emoji) => (
                  <button
                    key={emoji}
                    type="button"
                    onClick={() => handleEmojiSelect(emoji)}
                    className="flex size-8 items-center justify-center rounded-lg text-lg transition-colors hover:bg-muted"
                  >
                    {emoji}
                  </button>
                ))}
              </div>
            </div>
          )}

          <div className="overflow-hidden rounded-2xl bg-[var(--background)] ring-1 ring-[var(--border)] shadow-sm transition-shadow focus-within:ring-2 focus-within:ring-primary/20 focus-within:shadow-md">
            <textarea
              ref={innerRef}
              value={text}
              onChange={handleChange}
              onKeyDown={handleKeyDown}
              placeholder={
                disabled ? "请先选择联系人" : "输入消息… (Enter 发送，Shift+Enter 换行)"
              }
              disabled={disabled || uploading}
              rows={2}
              className="block w-full resize-none bg-transparent px-4 pt-3 text-sm text-foreground placeholder:text-muted-foreground/50 focus:outline-none disabled:cursor-not-allowed disabled:opacity-50"
              style={{ minHeight: "4rem", maxHeight: "120px" }}
            />

            <div className="flex items-center gap-4 px-4 pb-3 pt-1">
              <button
                type="button"
                title="表情"
                onClick={() => setEmojiOpen((prev) => !prev)}
                className={emojiOpen
                  ? "flex size-8 items-center justify-center rounded-full text-primary transition-colors"
                  : "flex size-8 items-center justify-center rounded-full text-muted-foreground/50 transition-colors hover:bg-muted-foreground/10 hover:text-muted-foreground"
                }
              >
                <Smile className="size-[18px]" />
              </button>
              <button
                type="button"
                title="文件"
                disabled={uploading}
                onClick={() => fileInputRef.current?.click()}
                className="flex size-8 items-center justify-center rounded-full text-muted-foreground/50 transition-colors hover:bg-muted-foreground/10 hover:text-muted-foreground disabled:opacity-50"
              >
                <Paperclip className="size-[18px]" />
              </button>
              <button
                type="button"
                title="图片"
                disabled={uploading}
                onClick={() => imageInputRef.current?.click()}
                className="flex size-8 items-center justify-center rounded-full text-muted-foreground/50 transition-colors hover:bg-muted-foreground/10 hover:text-muted-foreground disabled:opacity-50"
              >
                {uploading ? <Loader2 className="size-[18px] animate-spin" /> : <ImageIcon className="size-[18px]" />}
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  },
);

export { ChatInput };
