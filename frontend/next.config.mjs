/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',
  images: { unoptimized: true },
  allowedDevOrigins: ["192.168.0.130"],
};

export default nextConfig;
