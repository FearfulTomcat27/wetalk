"use client";

import { useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useAuthStore } from "@/stores/auth";

const features = ["即时消息，秒级送达", "安全可靠，隐私保护", "简洁易用，开箱即聊"];

export default function HomePage() {
  const router = useRouter();
  const { token, _hydrated } = useAuthStore();
  const init = useAuthStore((s) => s.init);
  const isLoggedIn = Boolean(token);

  // 客户端 hydration：从 localStorage 恢复 token
  useEffect(() => {
    init();
  }, [init]);

  // hydration 完成后 → 已登录则跳转
  useEffect(() => {
    if (_hydrated && isLoggedIn) {
      router.replace("/chat");
    }
  }, [_hydrated, isLoggedIn, router]);

  // SSR 或尚未 hydration → 等待
  if (!_hydrated) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black px-4">
        <p className="text-sm text-muted-foreground">加载中…</p>
      </div>
    );
  }

  // 已登录 → 显示跳转 loading（useEffect 中执行 replace）
  if (isLoggedIn) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black px-4">
        <p className="text-sm text-muted-foreground">正在跳转...</p>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-zinc-50 dark:bg-black px-4">
      <div className="w-full max-w-lg text-center space-y-8">
        {/* Hero */}
        <div className="space-y-4">
          <h1 className="text-4xl font-bold tracking-tight text-foreground sm:text-5xl">
            WeTalk
          </h1>
          <p className="text-xl text-muted-foreground">
            随时随地，畅快聊天
          </p>
        </div>

        {/* Feature List */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">功能特性</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="space-y-2">
              {features.map((text) => (
                <li key={text} className="text-sm text-muted-foreground">
                  {text}
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>

        {/* CTA Buttons */}
        <div className="flex flex-col gap-3 sm:flex-row sm:justify-center">
          <Button size="lg" asChild className="sm:w-32">
            <Link href="/login">登录</Link>
          </Button>
          <Button size="lg" variant="outline" asChild className="sm:w-32">
            <Link href="/register">注册</Link>
          </Button>
        </div>
      </div>
    </div>
  );
}
