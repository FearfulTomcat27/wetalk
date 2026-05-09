/** @type {import('next').NextConfig} */
const nextConfig = {
  images: {
    remotePatterns: [
      { protocol: 'https', hostname: 'api.dicebear.com' },
      { protocol: 'https', hostname: '*.oss-cn-shanghai.aliyuncs.com' },
      { protocol: 'http', hostname: 'localhost', port: '8080' },
      { protocol: 'http', hostname: '192.168.0.130', port: '8080' },
    ],
  },
  allowedDevOrigins:["192.168.0.130"]
};

export default nextConfig;
