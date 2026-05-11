"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { Copy, Trash2, Reply } from "lucide-react";
import { toast } from "sonner";
import type { Message } from "@/types/chat";
import { cn } from "@/lib/utils";

interface MessageContextMenuProps {
  x: number;
  y: number;
  message: Message;
  isSelf: boolean;
  onClose: () => void;
  onCopy?: (message: Message) => void;
  onDelete?: (message: Message) => void;
  onQuote?: (message: Message) => void;
}

export function MessageContextMenu({
  x,
  y,
  message,
  isSelf,
  onClose,
  onCopy,
  onDelete,
  onQuote,
}: MessageContextMenuProps) {
  const menuRef = useRef<HTMLDivElement>(null);
  const [adjustedX, setAdjustedX] = useState(x);
  const [adjustedY, setAdjustedY] = useState(y);

  // 根据 viewport 边界调整菜单位置（打开后执行一次）
  useEffect(() => {
    if (!menuRef.current) {return;}
    const rect = menuRef.current.getBoundingClientRect();
    let newX = x;
    let newY = y;

    if (x + rect.width > window.innerWidth) {
      newX = x - rect.width;
    }
    if (y + rect.height > window.innerHeight) {
      newY = y - rect.height;
    }

    if (newX !== x) {setAdjustedX(newX);}
    if (newY !== y) {setAdjustedY(newY);}
  }, [x, y]);

  // 点击外部关闭（mousedown，延迟注册避免右键事件立即触发关闭）
  const handleClickOutside = useCallback(
    (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        onClose();
      }
    },
    [onClose],
  );

  useEffect(() => {
    const timer = setTimeout(() => {
      document.addEventListener("mousedown", handleClickOutside);
      document.addEventListener("contextmenu", handleClickOutside);
    }, 0);
    return () => {
      clearTimeout(timer);
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("contextmenu", handleClickOutside);
    };
  }, [handleClickOutside]);

  // Escape 关闭
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") {onClose();}
    }
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [onClose]);

  const handleCopy = () => {
    navigator.clipboard.writeText(message.content).then(() => {
      toast.success("已复制");
    }).catch(() => {
      toast.error("复制失败");
    });
    onCopy?.(message);
    onClose();
  };

  const handleDelete = () => {
    onDelete?.(message);
    onClose();
  };

  const handleQuote = () => {
    if (onQuote) {
      onQuote(message);
    } else {
      toast.info("引用功能暂不可用");
    }
    onClose();
  };

  const isText = !message.content_type || message.content_type === "text";

  return (
    <div
      ref={menuRef}
      className="fixed z-50 min-w-[140px] rounded-lg border bg-popover p-1 text-popover-foreground shadow-lg"
      style={{ left: adjustedX, top: adjustedY }}
    >
      {isText && (
        <MenuItem onClick={handleCopy} icon={<Copy className="size-4" />}>
          复制
        </MenuItem>
      )}
      <MenuItem onClick={handleQuote} icon={<Reply className="size-4" />}>
        引用
      </MenuItem>
      {isSelf && (
        <>
          <div className="mx-1 my-1 h-px bg-border" />
          <MenuItem
            onClick={handleDelete}
            icon={<Trash2 className="size-4" />}
            destructive
          >
            删除
          </MenuItem>
        </>
      )}
    </div>
  );
}

function MenuItem({
  onClick,
  icon,
  destructive,
  children,
}: {
  onClick: () => void;
  icon: React.ReactNode;
  destructive?: boolean;
  children: React.ReactNode;
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "flex w-full cursor-pointer items-center gap-2.5 rounded-md px-2.5 py-1.5 text-sm transition-colors",
        destructive
          ? "text-destructive hover:bg-destructive/10"
          : "hover:bg-accent hover:text-accent-foreground",
      )}
    >
      {icon}
      {children}
    </button>
  );
}
