import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Keep framework-generated agent instructions out of the repository.
  agentRules: false,
};

export default nextConfig;
