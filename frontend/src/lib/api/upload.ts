import { request } from "./request";

export interface UploadResult {
  url: string;
  content_type?: string;
  file_name: string;
  file_size: number;
  file_type: string;
}

/** 上传文件（图片或通用文件） */
export function uploadFile(file: File, type: "image" | "file") {
  const formData = new FormData();
  formData.append("file", file);
  formData.append("type", type);
  return request<UploadResult>({
    method: "POST",
    url: "/api/upload",
    data: formData,
    headers: { "Content-Type": "multipart/form-data" },
  });
}