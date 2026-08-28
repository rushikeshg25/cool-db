"use client";

import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect } from "react";
import {
  VariantCommandCenter,
  variantName as commandCenterName,
} from "./variant-command-center";
import {
  VariantNotebook,
  variantName as notebookName,
} from "./variant-notebook";
import {
  VariantTerminal,
  variantName as terminalName,
} from "./variant-terminal";

const variants = [
  { key: "A", name: commandCenterName, render: VariantCommandCenter },
  { key: "B", name: notebookName, render: VariantNotebook },
  { key: "C", name: terminalName, render: VariantTerminal },
] as const;

// Three local database dashboard variants, switchable via ?variant=, on the
// existing root route. This is intentionally a throwaway design prototype.
export function DemoDashboard() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();
  const requestedKey = searchParams.get("variant")?.toUpperCase();
  const currentIndex = Math.max(
    0,
    variants.findIndex((variant) => variant.key === requestedKey),
  );
  const current = variants[currentIndex];
  const CurrentVariant = current.render;

  const switchTo = useCallback(
    (nextIndex: number) => {
      const wrappedIndex = (nextIndex + variants.length) % variants.length;
      const params = new URLSearchParams(searchParams.toString());
      params.set("variant", variants[wrappedIndex].key);
      router.replace(`${pathname}?${params.toString()}`, { scroll: false });
    },
    [pathname, router, searchParams],
  );

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      const target = event.target as HTMLElement | null;
      if (
        target?.matches("input, textarea, select, [contenteditable='true']") ||
        (event.key !== "ArrowLeft" && event.key !== "ArrowRight")
      ) {
        return;
      }
      event.preventDefault();
      switchTo(currentIndex + (event.key === "ArrowRight" ? 1 : -1));
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [currentIndex, switchTo]);

  return (
    <>
      <CurrentVariant />
      <PrototypeSwitcher
        currentIndex={currentIndex}
        onPrevious={() => switchTo(currentIndex - 1)}
        onNext={() => switchTo(currentIndex + 1)}
      />
    </>
  );
}

function PrototypeSwitcher({
  currentIndex,
  onPrevious,
  onNext,
}: {
  currentIndex: number;
  onPrevious: () => void;
  onNext: () => void;
}) {
  if (process.env.NODE_ENV === "production") return null;
  const current = variants[currentIndex];

  return (
    <nav
      aria-label="Dashboard prototype variants"
      className="fixed bottom-5 left-1/2 z-50 flex -translate-x-1/2 items-center gap-1 rounded-full border border-white/15 bg-[#111] p-1.5 font-sans text-white shadow-2xl shadow-black/40"
    >
      <button
        aria-label="Previous variant"
        className="grid size-9 place-items-center rounded-full text-sm text-white/60 transition hover:bg-white/10 hover:text-white"
        onClick={onPrevious}
      >
        ←
      </button>
      <span className="min-w-40 px-3 text-center text-xs font-medium">
        {current.key} — {current.name}
      </span>
      <button
        aria-label="Next variant"
        className="grid size-9 place-items-center rounded-full text-sm text-white/60 transition hover:bg-white/10 hover:text-white"
        onClick={onNext}
      >
        →
      </button>
    </nav>
  );
}
