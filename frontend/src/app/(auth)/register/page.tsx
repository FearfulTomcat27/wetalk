"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { toast } from "@/hooks/use-toast";
import { useAuthStore } from "@/stores/auth";
import { registerSchema } from "@/lib/validators";
import type { RegisterInput } from "@/lib/validators";

export default function RegisterPage() {
  const router = useRouter();
  const { register, isLoading } = useAuthStore();
  const [form, setForm] = useState<RegisterInput>({
    username: "",
    password: "",
    confirmPassword: "",
    nickname: "",
  });
  const [errors, setErrors] = useState<Partial<Record<keyof RegisterInput, string>>>({});

  function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    const { name, value } = e.target;
    setForm((prev) => ({ ...prev, [name]: value }));
    if (errors[name as keyof RegisterInput]) {
      setErrors((prev) => ({ ...prev, [name]: undefined }));
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();

    const result = registerSchema.safeParse(form);
    if (!result.success) {
      const fieldErrors: Partial<Record<keyof RegisterInput, string>> = {};
      for (const issue of result.error.issues) {
        const key = issue.path[0] as keyof RegisterInput;
        if (!fieldErrors[key]) {
          fieldErrors[key] = issue.message;
        }
      }
      setErrors(fieldErrors);
      return;
    }

    try {
      await register(result.data);
      toast.success("注册成功");
      router.push("/chat");
    } catch {
      // 错误 toast 由 api 拦截器统一处理
    }
  }

  return (
    <Card className="w-full max-w-sm">
      <CardHeader>
        <CardTitle className="text-2xl">注册</CardTitle>
        <CardDescription>创建一个新账号加入 WeTalk</CardDescription>
      </CardHeader>
      <form onSubmit={handleSubmit}>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="username">用户名</Label>
            <Input
              id="username"
              name="username"
              placeholder="3-20位用户名"
              value={form.username}
              onChange={handleChange}
            />
            {errors.username && (
              <p className="text-sm text-destructive">{errors.username}</p>
            )}
          </div>
          <div className="space-y-2">
            <Label htmlFor="nickname">昵称</Label>
            <Input
              id="nickname"
              name="nickname"
              placeholder="可选，最多20位"
              value={form.nickname}
              onChange={handleChange}
            />
            {errors.nickname && (
              <p className="text-sm text-destructive">{errors.nickname}</p>
            )}
          </div>
          <div className="space-y-2">
            <Label htmlFor="password">密码</Label>
            <Input
              id="password"
              name="password"
              type="password"
              placeholder="至少6位密码"
              value={form.password}
              onChange={handleChange}
            />
            {errors.password && (
              <p className="text-sm text-destructive">{errors.password}</p>
            )}
          </div>
          <div className="space-y-2">
            <Label htmlFor="confirmPassword">确认密码</Label>
            <Input
              id="confirmPassword"
              name="confirmPassword"
              type="password"
              placeholder="再次输入密码"
              value={form.confirmPassword}
              onChange={handleChange}
            />
            {errors.confirmPassword && (
              <p className="text-sm text-destructive">{errors.confirmPassword}</p>
            )}
          </div>
        </CardContent>
        <CardFooter className="flex flex-col space-y-4">
          <Button type="submit" className="w-full" disabled={isLoading}>
            {isLoading ? "注册中..." : "注册"}
          </Button>
          <p className="text-sm text-muted-foreground text-center">
            已有账号？{" "}
            <Link href="/login" className="text-primary hover:underline">
              立即登录
            </Link>
          </p>
        </CardFooter>
      </form>
    </Card>
  );
}
