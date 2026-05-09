/** @type {import('next').NextConfig} */
const nextConfig = {
  images: {
    remotePatterns: [
      { protocol: 'https', hostname: 'api.dicebear.com' },
      { protocol: 'https', hostname: '*.oss-cn-shanghai.aliyuncs.com' },
      { protocol: 'http', hostname: 'localhost', port: '8080' },
    ],
  },
};

export default nextConfig;
