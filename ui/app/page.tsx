import { Suspense } from "react";
import { DemoDashboard } from "./_demo/demo-dashboard";
import { VariantCommandCenter } from "./_demo/variant-command-center";

export default function Home() {
  return (
    <Suspense fallback={<VariantCommandCenter />}>
      <DemoDashboard />
    </Suspense>
  );
}
