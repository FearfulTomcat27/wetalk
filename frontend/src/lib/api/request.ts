import axios, { AxiosError } from "axios";
import { toast } from "sonner";

const BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data?: T;
}

export class ApiError extends Error {
  code: number;
  constructor(code: number, message: string) {
    super(message);
    this.code = code;
    this.name = "ApiError";
  }
}

export const client = axios.create({
  baseURL: BASE_URL,
  timeout: 10000,
  headers: { "Content-Type": "application/json" },
});

// 请求拦截器：自动附加 JWT token
client.interceptors.request.use((config) => {
  if (typeof window !== "undefined") {
    const token = localStorage.getItem("token");
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});

// 响应拦截器：统一错误处理 + toast 通知
client.interceptors.response.use(
  (res) => res,
  (error: AxiosError<ApiResponse>) => {
    if (typeof window !== "undefined") {
      if (error.response?.data) {
        const { message } = error.response.data;
        toast.error(message || "请求失败");
      } else if (error.code === "ECONNABORTED") {
        toast.error("请求超时，请稍后重试");
      } else if (!error.response) {
        toast.error("网络异常，请稍后重试");
      } else {
        toast.error("请求失败，请稍后重试");
      }
    }

    if (error.response?.data) {
      const { code, message } = error.response.data;
      throw new ApiError(code || error.response.status, message || "请求失败");
    }
    throw new ApiError(0, error.message || "网络错误");
  },
);

export async function request<T = unknown>(
  config: Parameters<typeof client.request>[0],
): Promise<ApiResponse<T>> {
  const res = await client.request<ApiResponse<T>>(config);
  return res.data;
}
