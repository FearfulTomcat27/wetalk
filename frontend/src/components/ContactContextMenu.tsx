"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { Trash2, CheckCheck, Pin, Ellipsis } from "lucide-react";
import { cn } from "@/lib/utils";

interface ContactContextMenuProps {
  x: number;
  y: number;
  /** 联系人 id（chat 页面中的 contact.id） */
  contactId: number;
  /** 联系人昵称，用于提示文字 */
  contactName: string;
  onClose: () => void;
  onDeleteChatHistory?: (contactId: number) => void;
}

export function ContactContextMenu({
  x,
  y,
  contactId,
  contactName,
  onClose,
  onDeleteChatHistory,
}: ContactContextMenuProps) {
  const menuRef = useRef<HTMLDivElement>(null);
  const [adjustedX, setAdjustedX] = useState(x);
  const [adjustedY, setAdjustedY] = useState(y);

  // 根据 viewport 边界调整菜单位置
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

  // 点击外部关闭
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

  const handleDelete = () => {
    onDeleteChatHistory?.(contactId);
    onClose();
  };

  return (
    <div
      ref={menuRef}
      className="fixed z-50 min-w-[160px] rounded-lg border bg-popover p-1 text-popover-foreground shadow-lg"
      style={{ left: adjustedX, top: adjustedY }}
    >
      <div className="px-2.5 py-1.5 text-xs text-muted-foreground truncate max-w-[180px]">
        {contactName}
      </div>
      <div className="mx-1 my-1 h-px bg-border" />

      {/* 标记已读 — 预留 */}
      <MenuItem icon={<CheckCheck className="size-4" />} disabled>
        标记已读
      </MenuItem>

      {/* 置顶聊天 — 预留 */}
      <MenuItem icon={<Pin className="size-4" />} disabled>
        置顶聊天
      </MenuItem>

      {/* 更多 — 预留 */}
      <MenuItem icon={<Ellipsis className="size-4" />} disabled>
        更多
      </MenuItem>

      <div className="mx-1 my-1 h-px bg-border" />

      <MenuItem
        onClick={handleDelete}
        icon={<Trash2 className="size-4" />}
        destructive
      >
        删除聊天记录
      </MenuItem>
    </div>
  );
}

function MenuItem({
  onClick,
  icon,
  destructive,
  disabled,
  children,
}: {
  onClick?: () => void;
  icon: React.ReactNode;
  destructive?: boolean;
  disabled?: boolean;
  children: React.ReactNode;
}) {
  return (
    <button
      onClick={disabled ? undefined : onClick}
      disabled={disabled}
      className={cn(
        "flex w-full cursor-pointer items-center gap-2.5 rounded-md px-2.5 py-1.5 text-sm transition-colors",
        disabled && "cursor-not-allowed opacity-40",
        destructive && !disabled
          ? "text-destructive hover:bg-destructive/10"
          : "hover:bg-accent hover:text-accent-foreground",
      )}
    >
      {icon}
      {children}
    </button>
  );
}
