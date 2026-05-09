"use client";

import { useRef, useState } from "react";
import Image from "next/image";
import { Popover, PopoverTrigger, PopoverContent } from "@/components/ui/popover";
import { Button } from "@/components/ui/button";
import { uploadAvatar } from "@/lib/api";
import { getAvatarSrc } from "@/lib/avatar";
import { useAuthStore } from "@/stores/auth";
import { Loader2, Camera } from "lucide-react";
import { toast } from "sonner";

const MAX_FILE_SIZE = 2 * 1024 * 1024; // 2MB
const ACCEPTED_TYPES = ["image/jpeg", "image/png", "image/gif", "image/webp"];

interface ProfilePopoverProps {
  children: React.ReactNode;
}

export function ProfilePopover({ children }: ProfilePopoverProps) {
  const user = useAuthStore((s) => s.user);
  const updateAvatar = useAuthStore((s) => s.updateAvatar);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [open, setOpen] = useState(false);

  const username = user?.username ?? "me";
  const currentAvatarSrc = getAvatarSrc(user?.avatar, username);

  function handleFileSelect(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) {
      return;
    }

    if (!ACCEPTED_TYPES.includes(file.type)) {
      toast.error("仅支持 JPG、PNG、GIF、WebP 格式");
      e.target.value = "";
      return;
    }

    if (file.size > MAX_FILE_SIZE) {
      toast.error("图片大小不能超过 2MB");
      e.target.value = "";
      return;
    }

    setSelectedFile(file);
    const url = URL.createObjectURL(file);
    setPreviewUrl(url);
  }

  function handleAvatarClick() {
    fileInputRef.current?.click();
  }

  async function handleUpload() {
    if (!selectedFile) {
      return;
    }

    setUploading(true);
    try {
      const res = await uploadAvatar(selectedFile);
      const newAvatarUrl = res.data!.avatar;
      updateAvatar(newAvatarUrl);
      toast.success("头像更新成功");
      cleanup();
      setOpen(false);
    } catch {
      // 错误已由 axios 拦截器 toast 处理
    } finally {
      setUploading(false);
    }
  }

  function cleanup() {
    if (previewUrl) {
      URL.revokeObjectURL(previewUrl);
    }
    setPreviewUrl(null);
    setSelectedFile(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  }

  function handleOpenChange(val: boolean) {
    setOpen(val);
    if (!val) {
      cleanup();
    }
  }

  const displaySrc = previewUrl || currentAvatarSrc;

  return (
    <Popover open={open} onOpenChange={handleOpenChange}>
      <PopoverTrigger asChild>{children}</PopoverTrigger>
      <PopoverContent align="start" side="right" className="w-64 p-4">
        <div className="flex flex-col items-center gap-3">
          {/* 大头像 */}
          <div
            onClick={handleAvatarClick}
            className="group relative size-20 cursor-pointer rounded-lg"
          >
            <Image
              src={displaySrc}
              alt={user?.nickname || user?.username || ""}
              width={80}
              height={80}
              unoptimized={displaySrc.includes("dicebear")}
              className="rounded-lg object-cover ring-2 ring-border"
            />
            {/* 点击提示蒙层 */}
            <div className="absolute inset-0 flex items-center justify-center rounded-lg bg-black/40 opacity-0 transition-opacity group-hover:opacity-100">
              <Camera className="size-6 text-white" />
            </div>
            {uploading && (
              <div className="absolute inset-0 flex items-center justify-center rounded-lg bg-black/50">
                <Loader2 className="size-6 animate-spin text-white" />
              </div>
            )}
          </div>

          <input
            ref={fileInputRef}
            type="file"
            accept={ACCEPTED_TYPES.join(",")}
            onChange={handleFileSelect}
            className="hidden"
          />

          {/* 用户名 */}
          <div className="text-center">
            <p className="text-sm font-semibold text-foreground">
              {user?.nickname || user?.username}
            </p>
            <p className="text-xs text-muted-foreground">@{user?.username}</p>
          </div>

          {/* 确认上传按钮（仅在选择新图片后显示） */}
          {previewUrl && (
            <div className="flex gap-2">
              <Button
                size="sm"
                onClick={handleUpload}
                disabled={uploading}
              >
                {uploading ? (
                  <>
                    <Loader2 className="size-4 animate-spin" />
                    上传中...
                  </>
                ) : (
                  "确认上传"
                )}
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={cleanup}
                disabled={uploading}
              >
                取消
              </Button>
            </div>
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
}