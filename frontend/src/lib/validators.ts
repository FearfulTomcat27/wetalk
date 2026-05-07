import { z } from "zod";

export const loginSchema = z.object({
  username: z
    .string()
    .min(1, "请输入用户名"),
  password: z
    .string()
    .min(1, "请输入密码"),
});

export const registerSchema = z
  .object({
    username: z
      .string()
      .min(3, "用户名至少3位")
      .max(20, "用户名最多20位"),
    password: z
      .string()
      .min(6, "密码至少6位"),
    confirmPassword: z
      .string()
      .min(1, "请确认密码"),
    nickname: z
      .string()
      .max(20, "昵称最多20位")
      .optional()
      .or(z.literal("")),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "两次密码不一致",
    path: ["confirmPassword"],
  });

export type LoginInput = z.infer<typeof loginSchema>;
export type RegisterInput = z.infer<typeof registerSchema>;
