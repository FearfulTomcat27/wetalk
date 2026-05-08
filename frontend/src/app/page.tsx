"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useAuthStore } from "@/stores/auth";

const features = ["即时消息，秒级送达", "安全可靠，隐私保护", "简洁易用，开箱即聊"];

/**
 * Dashboard 首页 — 纯展示组件。
 * 已登录用户由 RouteGuard 自动跳转 /chat，此处仅渲染未登录着陆页。
 */
export default function HomePage() {
  const token = useAuthStore((s) => s.token);

  // 已登录 → RouteGuard 会处理跳转和 loading UI
  if (token) {
    return null;
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-zinc-50 dark:bg-black px-4">
      <div className="w-full max-w-lg text-center space-y-8">
        <div className="space-y-4">
          <h1 className="text-4xl font-bold tracking-tight text-foreground sm:text-5xl">
            WeTalk
          </h1>
          <p className="text-xl text-muted-foreground">
            随时随地，畅快聊天
          </p>
        </div>

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
