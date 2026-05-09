"use client";

import { useState, useCallback, useRef } from "react";
import { Search, UserPlus, Loader2 } from "lucide-react";
import { Avatar } from "@/components/Avatar";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { searchUsers, addFriend, ApiError } from "@/lib/api";
import type { UserInfo } from "@/lib/api";

interface AddFriendDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AddFriendDialog({ open, onOpenChange }: AddFriendDialogProps) {
  const [keyword, setKeyword] = useState("");
  const [results, setResults] = useState<UserInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);
  const [addingIds, setAddingIds] = useState<Set<number>>(new Set());
  const [addedIds, setAddedIds] = useState<Set<number>>(new Set());
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const handleSearch = useCallback(
    (value: string) => {
      setKeyword(value);

      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }

      if (!value.trim()) {
        setResults([]);
        setSearched(false);
        setAddedIds(new Set());
        return;
      }

      timerRef.current = setTimeout(async () => {
        setLoading(true);
        setSearched(true);
        setAddedIds(new Set());
        try {
          const res = await searchUsers(value.trim());
          setResults(res.data || []);
        } catch (err) {
          if (!(err instanceof ApiError)) {
            toast.error("搜索失败，请稍后重试");
          }
          setResults([]);
        } finally {
          setLoading(false);
        }
      }, 300);
    },
    [],
  );

  const handleAddFriend = useCallback(async (userId: number) => {
    setAddingIds((prev) => {
      const next = new Set(prev);
      next.add(userId);
      return next;
    });
    try {
      await addFriend(userId);
      toast.success("好友请求已发送");
      setAddedIds((prev) => {
        const next = new Set(prev);
        next.add(userId);
        return next;
      });
    } catch {
      // 错误已在拦截器中 toast
    } finally {
      setAddingIds((prev) => {
        const next = new Set(prev);
        next.delete(userId);
        return next;
      });
    }
  }, []);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>添加好友</DialogTitle>
          <DialogDescription>
            输入对方的用户名进行搜索
          </DialogDescription>
        </DialogHeader>

        {/* 搜索框 */}
        <div className="relative">
          <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground/50" />
          <input
            type="text"
            value={keyword}
            onChange={(e) => handleSearch(e.target.value)}
            placeholder="输入用户名搜索…"
            autoFocus
            className="w-full rounded-lg border bg-muted/50 py-2.5 pl-9 pr-4 text-sm outline-none transition-colors focus:border-primary/30 focus:bg-background focus:ring-2 focus:ring-primary/10"
          />
        </div>

        {/* 搜索结果 */}
        <div className="max-h-64 overflow-y-auto">
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="size-5 animate-spin text-muted-foreground" />
            </div>
          ) : searched && results.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-8 text-center">
              <Search className="mb-2 size-8 text-muted-foreground/25" />
              <p className="text-sm text-muted-foreground">未找到用户</p>
              <p className="mt-1 text-xs text-muted-foreground/50">
                试试其他关键词
              </p>
            </div>
          ) : (
            <ul className="space-y-1 py-1">
              {results.map((user) => (
                <li
                  key={user.id}
                  className="flex items-center gap-3 rounded-lg px-3 py-2.5 transition-colors hover:bg-muted/60"
                >
                  {/* 头像 */}
                  <Avatar
                    src={user.avatar}
                    alt={user.nickname || user.username}
                    size={36}
                  />

                  {/* 信息 */}
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium">
                      {user.nickname}
                    </p>
                    <p className="truncate text-xs text-muted-foreground">
                      @{user.username}
                    </p>
                  </div>

                  {/* 添加按钮 */}
                  {addedIds.has(user.id) ? (
                    <span className="shrink-0 rounded-lg bg-muted px-3 py-1.5 text-xs font-medium text-muted-foreground">
                      已发送
                    </span>
                  ) : (
                    <button
                      onClick={() => handleAddFriend(user.id)}
                      disabled={addingIds.has(user.id)}
                      className="inline-flex shrink-0 items-center gap-1.5 rounded-lg bg-primary px-3 py-1.5 text-xs font-medium text-primary-foreground transition-all hover:bg-primary/90 active:scale-95 disabled:cursor-not-allowed disabled:opacity-50"
                    >
                      {addingIds.has(user.id) ? (
                        <Loader2 className="size-3.5 animate-spin" />
                      ) : (
                        <UserPlus className="size-3.5" />
                      )}
                      添加
                    </button>
                  )}
                </li>
              ))}
            </ul>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
