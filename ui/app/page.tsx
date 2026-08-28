import { VariantCommandCenter } from "./_demo/variant-command-center";

// Three local dashboard variants will live on this route, switchable with
// ?variant=. Variant A lands first so every prototype commit remains runnable.
export default function Home() {
  return <VariantCommandCenter />;
}
