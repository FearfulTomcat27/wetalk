"use client";

import { useState, useRef, useCallback, useEffect, type MouseEvent as ReactMouseEvent } from "react";
import { useRouter } from "next/navigation";
import { ContactList } from "@/components/ContactList";
import { Avatar } from "@/components/Avatar";
import type { Contact } from "@/types/chat";
import { MessageCircle, UserPlus, Loader2, Check, X, Users } from "lucide-react";
import { cn } from "@/lib/utils";
import { getFriends, getPendingRequests, acceptFriendRequest, declineFriendRequest } from "@/lib/api";
import type { PendingRequest, FriendInfo } from "@/lib/api";
import { toast } from "sonner";

// 宽度常量
const DEFAULT_CONTACT_WIDTH = 280;
const MIN_CONTACT_WIDTH = 200;
const MAX_CONTACT_WIDTH = 480;
const MIN_DETAIL_WIDTH = 300;

/** 将后端 FriendInfo 映射为前端 Contact 类型 */
function mapFriendToContact(f: FriendInfo): Contact {
  return {
    id: f.friend_id,
    username: f.friend_name,
    nickname: f.friend_name,
    avatar: f.friend_avatar,
    unread: 0,
  };
}

export default function ContactsPage() {
  const router = useRouter();
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [contactsLoading, setContactsLoading] = useState(true);
  const [activeContactId, setActiveContactId] = useState<number | null>(null);

  // 新的朋友
  const [showPending, setShowPending] = useState(false);
  const [pendingRequests, setPendingRequests] = useState<PendingRequest[]>([]);
  const [pendingLoading, setPendingLoading] = useState(false);
  const [processingIds, setProcessingIds] = useState<Set<number>>(new Set());
  const pendingCount = pendingRequests.length;

  // 拖拽
  const [contactWidth, setContactWidth] = useState(DEFAULT_CONTACT_WIDTH);
  const [dragging, setDragging] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const activeContact = contacts.find((c) => c.id === activeContactId) ?? null;

  // 加载好友列表
  useEffect(() => {
    async function load() {
      setContactsLoading(true);
      try {
        const res = await getFriends();
        const mapped: Contact[] = (res.data || []).map(mapFriendToContact);
        setContacts(mapped);
      } catch {
        // 错误已在拦截器 toast
      } finally {
        setContactsLoading(false);
      }
    }
    load();
  }, []);

  // 点击「新的朋友」
  async function handleNewFriends() {
    setActiveContactId(null);
    setShowPending(true);
    setPendingLoading(true);
    try {
      const res = await getPendingRequests();
      setPendingRequests(res.data || []);
    } catch {
      // 错误已在拦截器 toast
    } finally {
      setPendingLoading(false);
    }
  }

  // 处理好友请求
  async function handleAccept(reqId: number) {
    setProcessingIds((prev) => {
      const next = new Set(prev);
      next.add(reqId);
      return next;
    });
    try {
      await acceptFriendRequest(reqId);
      toast.success("已同意好友请求");
      setPendingRequests((prev) => prev.filter((r) => r.id !== reqId));
      // 刷新好友列表
      const fres = await getFriends();
      setContacts((fres.data || []).map(mapFriendToContact));
    } catch {
      // 错误已在拦截器 toast
    } finally {
      setProcessingIds((prev) => {
        const next = new Set(prev);
        next.delete(reqId);
        return next;
      });
    }
  }

  async function handleDecline(reqId: number) {
    setProcessingIds((prev) => {
      const next = new Set(prev);
      next.add(reqId);
      return next;
    });
    try {
      await declineFriendRequest(reqId);
      toast.success("已拒绝好友请求");
      setPendingRequests((prev) => prev.filter((r) => r.id !== reqId));
    } catch {
      // 错误已在拦截器 toast
    } finally {
      setProcessingIds((prev) => {
        const next = new Set(prev);
        next.delete(reqId);
        return next;
      });
    }
  }

  // 拖拽处理
  const handleMouseDown = useCallback((e: ReactMouseEvent) => {
    e.preventDefault();
    setDragging(true);
  }, []);

  useEffect(() => {
    if (!dragging) {return;}

    function handleMouseMove(e: globalThis.MouseEvent) {
      if (!containerRef.current) {return;}
      const rect = containerRef.current.getBoundingClientRect();
      const newWidth = e.clientX - rect.left;
      const maxWidth = Math.min(rect.width - MIN_DETAIL_WIDTH, MAX_CONTACT_WIDTH);
      setContactWidth(Math.min(Math.max(newWidth, MIN_CONTACT_WIDTH), maxWidth));
    }

    function handleMouseUp() {
      setDragging(false);
    }

    document.addEventListener("mousemove", handleMouseMove);
    document.addEventListener("mouseup", handleMouseUp);
    document.body.style.userSelect = "none";
    document.body.style.cursor = "col-resize";

    return () => {
      document.removeEventListener("mousemove", handleMouseMove);
      document.removeEventListener("mouseup", handleMouseUp);
      document.body.style.userSelect = "";
      document.body.style.cursor = "";
    };
  }, [dragging]);

  function handleSendMessage() {
    if (activeContactId !== null) {
      router.push(`/chat?contact=${activeContactId}`);
    }
  }

  function handleSelectContact(id: number) {
    setShowPending(false);
    setActiveContactId(id);
  }

  // 「新的朋友」入口
  const newFriendsHeader = (
    <button
      onClick={handleNewFriends}
      className={cn(
        "mx-2 flex items-center gap-3 rounded-lg px-3 py-2.5 text-left transition-colors",
        showPending ? "bg-[#3b82f6] text-white" : "hover:bg-muted/70 text-foreground",
      )}
    >
      <div className={cn("flex size-[36px] shrink-0 items-center justify-center rounded-lg shadow-[0_2px_6px_rgba(0,0,0,0.14)]", showPending ? "bg-[#3b82f6] text-white" : "bg-[#3b82f6]/15 text-[#3b82f6]")}>
        <UserPlus className="size-5" />
      </div>
      <div className="min-w-0 flex-1">
        <p className="text-sm font-medium">
          新的朋友
        </p>
        {pendingCount > 0 && (
          <span className="mt-0.5 inline-block rounded-full bg-destructive px-1.5 py-0 text-[10px] font-bold text-destructive-foreground">
            {pendingCount > 99 ? "99+" : pendingCount}
          </span>
        )}
      </div>
    </button>
  );

  return (
    <div ref={containerRef} className="flex flex-1 overflow-hidden">
      {/* 联系人列表 + 新的朋友 */}
      {contactsLoading ? (
        <div className="flex items-center justify-center" style={{ width: contactWidth }}>
          <Loader2 className="size-6 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <ContactList
          contacts={contacts}
          activeContactId={activeContactId}
          onSelectContact={handleSelectContact}
          style={{ width: contactWidth }}
          variant="contacts"
          headerContent={newFriendsHeader}
        />
      )}

      {/* 拖拽手柄 */}
      <div
        onMouseDown={handleMouseDown}
        className={cn(
          "relative shrink-0 cursor-col-resize transition-colors",
          dragging ? "bg-primary/10" : "hover:bg-primary/30",
        )}
      >
        <div
          className={cn(
            "absolute inset-y-0 left-1/2 w-px -translate-x-1/2 transition-colors",
            dragging ? "bg-primary" : "bg-border",
          )}
        />
      </div>

      {/* 右侧面板 */}
      <div className="flex flex-1 flex-col overflow-hidden bg-muted/30">
        {/* 待处理请求列表 */}
        {showPending ? (
          <div className="flex flex-1 flex-col overflow-y-auto">
            <div className="flex h-14 shrink-0 items-center bg-card px-4">
              <h2 className="text-sm font-semibold">新的朋友</h2>
            </div>
            <div className="flex-1 overflow-y-auto">
              {pendingLoading ? (
                <div className="flex items-center justify-center py-16">
                  <Loader2 className="size-5 animate-spin text-muted-foreground" />
                </div>
              ) : pendingRequests.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-16 text-center">
                  <Users className="mb-3 size-10 text-muted-foreground/25" />
                  <p className="text-sm text-muted-foreground">暂无待处理的好友请求</p>
                </div>
              ) : (
                <ul className="space-y-0.5 p-2">
                  {pendingRequests.map((req) => (
                    <li
                      key={req.id}
                      className="flex items-center gap-3 rounded-lg px-3 py-3 transition-colors hover:bg-muted/50"
                    >
                      {/* 头像 */}
                      <Avatar
                        src={req.user.avatar}
                        alt={req.user.nickname || req.user.username}
                        size={36}
                        className="ring-2 ring-border"
                      />

                      {/* 信息 */}
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm font-medium">{req.user.nickname}</p>
                        <p className="truncate text-xs text-muted-foreground">
                          @{req.user.username}
                        </p>
                      </div>

                      {/* 操作按钮 */}
                      {processingIds.has(req.id) ? (
                        <Loader2 className="size-5 animate-spin text-muted-foreground" />
                      ) : (
                        <div className="flex shrink-0 items-center gap-1.5">
                          <button
                            onClick={() => handleAccept(req.id)}
                            className="inline-flex items-center gap-1 rounded-lg bg-primary px-3 py-1.5 text-xs font-medium text-primary-foreground transition-colors hover:bg-primary/90"
                          >
                            <Check className="size-3.5" />
                            同意
                          </button>
                          <button
                            onClick={() => handleDecline(req.id)}
                            className="inline-flex items-center gap-1 rounded-lg border bg-background px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive hover:border-destructive/30"
                          >
                            <X className="size-3.5" />
                            拒绝
                          </button>
                        </div>
                      )}
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </div>
        ) : activeContact ? (
          <div className="flex flex-1 flex-col items-center justify-center gap-6 px-6 py-12">
            {/* 头像 */}
            <Avatar
              src={activeContact.avatar}
              alt={activeContact.nickname || activeContact.username}
              size={36}
              className="ring-4 ring-border"
            />

            {/* 信息 */}
            <div className="text-center">
              <h2 className="text-xl font-semibold text-foreground">
                {activeContact.nickname}
              </h2>
              <p className="mt-1 text-sm text-muted-foreground">
                @{activeContact.username}
              </p>
            </div>

            {/* 发送消息按钮 */}
            <button
              onClick={handleSendMessage}
              className="inline-flex items-center gap-2 rounded-xl bg-primary px-6 py-2.5 text-sm font-medium text-primary-foreground shadow-sm transition-all hover:bg-primary/90 active:scale-95"
            >
              <MessageCircle className="size-4" />
              发送消息
            </button>
          </div>
        ) : (
          <div className="flex flex-1 items-center justify-center">
            <div className="text-center">
              <div className="mx-auto mb-4 flex size-20 items-center justify-center rounded-full bg-muted">
                <Users className="size-10 text-muted-foreground/40" />
              </div>
              <p className="text-lg font-medium text-muted-foreground">
                选择一个联系人查看详情
              </p>
              <p className="mt-1 text-sm text-muted-foreground/60">
                从左侧列表中选择一个联系人
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
