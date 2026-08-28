"use client";

import { demoQueries, useDemoQuery } from "./use-demo-query";

export const variantName = "Command center";

export function VariantCommandCenter() {
  const demo = useDemoQuery();
  const statusColor =
    demo.connection === "online"
      ? "bg-emerald-400"
      : demo.connection === "offline"
        ? "bg-rose-400"
        : "bg-amber-300";

  return (
    <main className="min-h-screen bg-[#0b0e0d] text-[#e6ece8] selection:bg-emerald-300 selection:text-black">
      <header className="flex h-16 items-center justify-between border-b border-white/8 px-5 lg:px-7">
        <div className="flex items-center gap-3">
          <div className="grid size-9 place-items-center rounded-xl bg-emerald-300 font-mono text-sm font-black text-[#07100b] shadow-[0_0_30px_rgba(110,231,183,.15)]">
            C/
          </div>
          <div>
            <p className="text-sm font-semibold tracking-tight">CoolDB Studio</p>
            <p className="font-mono text-[10px] uppercase tracking-[.2em] text-white/35">
              local prototype
            </p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <button
            className="hidden rounded-lg border border-white/10 px-3 py-2 text-xs text-white/55 transition hover:border-white/20 hover:text-white sm:block"
            onClick={() => void demo.checkConnection()}
          >
            Reconnect
          </button>
          <div className="flex items-center gap-2 rounded-full border border-white/10 bg-white/[.03] px-3 py-2 font-mono text-[11px] text-white/60">
            <span className={`size-1.5 rounded-full ${statusColor}`} />
            {demo.connection}
          </div>
        </div>
      </header>

      <div className="grid min-h-[calc(100vh-4rem)] lg:grid-cols-[230px_minmax(0,1fr)_250px]">
        <aside className="hidden border-r border-white/8 p-5 lg:block">
          <p className="mb-4 text-[10px] font-semibold uppercase tracking-[.22em] text-white/30">
            Explorer
          </p>
          <div className="rounded-xl border border-white/8 bg-white/[.025] p-3">
            <div className="flex items-center gap-2 text-xs font-medium">
              <span className="text-emerald-300">◆</span> default.cooldb
            </div>
            <div className="ml-2 mt-4 border-l border-white/10 pl-4">
              <p className="mb-3 text-xs text-white/70">users</p>
              {[
                ["id", "INTEGER · PK"],
                ["name", "TEXT"],
                ["active", "BOOLEAN"],
              ].map(([name, type]) => (
                <div className="mb-2 flex justify-between gap-2" key={name}>
                  <span className="font-mono text-[11px] text-white/55">{name}</span>
                  <span className="text-[9px] text-white/25">{type}</span>
                </div>
              ))}
            </div>
          </div>
          <p className="mb-3 mt-8 text-[10px] font-semibold uppercase tracking-[.22em] text-white/30">
            Quick queries
          </p>
          <div className="space-y-1.5">
            {demoQueries.map((example) => (
              <button
                className="w-full rounded-lg px-3 py-2 text-left text-xs text-white/45 transition hover:bg-white/5 hover:text-white/80"
                key={example.label}
                onClick={() => demo.setQuery(example.query)}
              >
                <span className="mr-2 text-emerald-300/60">›</span>
                {example.label}
              </button>
            ))}
          </div>
        </aside>

        <section className="min-w-0 p-4 sm:p-7 lg:p-9">
          <div className="mx-auto max-w-4xl">
            <div className="mb-7 flex items-end justify-between gap-4">
              <div>
                <p className="mb-2 font-mono text-[10px] uppercase tracking-[.2em] text-emerald-300/65">
                  workspace / query 01
                </p>
                <h1 className="text-2xl font-semibold tracking-[-.035em] sm:text-3xl">
                  Query your local database
                </h1>
              </div>
              <span className="hidden text-xs text-white/30 sm:block">⌘ + Enter to run</span>
            </div>

            <div className="overflow-hidden rounded-2xl border border-white/10 bg-[#101413] shadow-2xl shadow-black/25">
              <div className="flex items-center justify-between border-b border-white/8 px-4 py-3">
                <div className="flex gap-1.5">
                  <span className="size-2 rounded-full bg-rose-400/70" />
                  <span className="size-2 rounded-full bg-amber-300/70" />
                  <span className="size-2 rounded-full bg-emerald-300/70" />
                </div>
                <span className="font-mono text-[10px] text-white/25">query.sql</span>
              </div>
              <div className="grid grid-cols-[42px_1fr]">
                <div className="border-r border-white/5 py-5 text-center font-mono text-xs leading-7 text-white/18">
                  1
                </div>
                <textarea
                  aria-label="SQL query"
                  className="min-h-44 resize-y bg-transparent p-5 font-mono text-[13px] leading-7 text-emerald-50 outline-none placeholder:text-white/20"
                  onChange={(event) => demo.setQuery(event.target.value)}
                  onKeyDown={(event) => {
                    if ((event.metaKey || event.ctrlKey) && event.key === "Enter") {
                      event.preventDefault();
                      void demo.runQuery();
                    }
                  }}
                  spellCheck={false}
                  value={demo.query}
                />
              </div>
              <div className="flex items-center justify-between border-t border-white/8 bg-black/10 px-4 py-3">
                <span className="font-mono text-[10px] text-white/25">SQL · UTF-8</span>
                <button
                  className="rounded-lg bg-emerald-300 px-4 py-2 text-xs font-bold text-[#07100b] transition hover:bg-emerald-200 disabled:cursor-wait disabled:opacity-50"
                  disabled={demo.running}
                  onClick={() => void demo.runQuery()}
                >
                  {demo.running ? "Running…" : "Run query ↗"}
                </button>
              </div>
            </div>

            <div className="mt-5 overflow-hidden rounded-2xl border border-white/8 bg-white/[.02]">
              <div className="flex items-center justify-between border-b border-white/8 px-4 py-3">
                <p className="text-xs font-medium">Result</p>
                <p className="font-mono text-[10px] text-white/30">
                  {demo.result.elapsedMs === null ? "waiting" : `${demo.result.elapsedMs} ms`}
                </p>
              </div>
              <pre
                className={`min-h-36 overflow-x-auto p-5 font-mono text-xs leading-6 ${demo.result.error ? "text-rose-300" : "text-white/65"}`}
              >
                {demo.result.error || demo.result.output}
              </pre>
            </div>
          </div>
        </section>

        <aside className="hidden border-l border-white/8 p-5 lg:block">
          <p className="mb-4 text-[10px] font-semibold uppercase tracking-[.22em] text-white/30">
            Session
          </p>
          <div className="space-y-3">
            {[
              ["Engine", "v0.1.0"],
              ["Transport", "HTTP → gRPC"],
              ["Database", "local"],
              ["Durability", "snapshot"],
            ].map(([label, value]) => (
              <div className="rounded-xl border border-white/8 bg-white/[.025] p-3" key={label}>
                <p className="text-[10px] uppercase tracking-wider text-white/25">{label}</p>
                <p className="mt-1.5 font-mono text-xs text-white/65">{value}</p>
              </div>
            ))}
          </div>
          <div className="mt-6 rounded-xl border border-amber-300/15 bg-amber-300/[.04] p-4">
            <p className="text-xs font-medium text-amber-200">Prototype mode</p>
            <p className="mt-2 text-[11px] leading-5 text-white/35">
              Local queries can mutate your demo database. Use a scratch file while exploring.
            </p>
          </div>
        </aside>
      </div>
    </main>
  );
}
