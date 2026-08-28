"use client";

import { demoQueries, useDemoQuery } from "./use-demo-query";

export const variantName = "Query notebook";

export function VariantNotebook() {
  const demo = useDemoQuery();

  return (
    <main className="min-h-screen bg-[#f3efe5] text-[#292720] selection:bg-[#ffcf70]">
      <header className="border-b border-[#292720]/10 bg-[#f7f3e9] px-5 py-4 md:px-10">
        <div className="mx-auto flex max-w-6xl items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="rounded-md border-2 border-[#292720] bg-[#ffcf70] px-2 py-1 font-mono text-sm font-black shadow-[3px_3px_0_#292720]">
              CDB
            </div>
            <div>
              <p className="font-serif text-lg font-bold tracking-tight">The CoolDB Notebook</p>
              <p className="text-[10px] uppercase tracking-[.18em] text-[#292720]/45">
                Friday · local exploration
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2 rounded-full border border-[#292720]/15 bg-white/55 px-3 py-2 text-xs">
            <span
              className={`size-2 rounded-full ${demo.connection === "online" ? "bg-[#26815d]" : "bg-[#cc5a46]"}`}
            />
            {demo.connection === "online" ? "Database ready" : "Database offline"}
          </div>
        </div>
      </header>

      <div className="mx-auto grid max-w-6xl gap-8 px-5 py-8 md:px-10 lg:grid-cols-[180px_1fr] lg:py-12">
        <aside className="hidden lg:block">
          <p className="mb-4 font-mono text-[10px] uppercase tracking-[.2em] text-[#292720]/40">
            Contents
          </p>
          <ol className="space-y-3 border-l border-[#292720]/15 pl-4 text-xs text-[#292720]/55">
            <li className="font-semibold text-[#292720]">01 / Scratch query</li>
            <li>02 / Output</li>
            <li>03 / Schema notes</li>
          </ol>
          <div className="mt-10 rotate-[-2deg] border border-[#d5b45d] bg-[#ffe6a6] p-4 shadow-sm">
            <p className="font-serif text-sm font-bold">Remember</p>
            <p className="mt-2 text-xs leading-5 text-[#292720]/65">
              Start the server with the local HTTP bridge on port 3041.
            </p>
          </div>
        </aside>

        <section className="min-w-0">
          <div className="mb-8 border-b-2 border-[#292720] pb-5">
            <p className="mb-2 font-mono text-[10px] uppercase tracking-[.22em] text-[#a34d39]">
              Experiment 01
            </p>
            <h1 className="font-serif text-4xl font-bold leading-tight tracking-[-.035em] md:text-5xl">
              What is inside this database?
            </h1>
            <p className="mt-4 max-w-2xl font-serif text-base italic leading-7 text-[#292720]/60">
              A small working notebook for creating tables, testing mutations, and reading local results.
            </p>
          </div>

          <div className="mb-4 flex flex-wrap gap-2">
            {demoQueries.map((example, index) => (
              <button
                className="rounded-full border border-[#292720]/15 bg-white/50 px-3 py-1.5 text-[11px] transition hover:-translate-y-0.5 hover:border-[#292720]/40 hover:bg-white"
                key={example.label}
                onClick={() => demo.setQuery(example.query)}
              >
                {String(index + 1).padStart(2, "0")} · {example.label}
              </button>
            ))}
          </div>

          <article className="overflow-hidden border-2 border-[#292720] bg-[#fffdf7] shadow-[7px_7px_0_rgba(41,39,32,.12)]">
            <div className="flex items-center justify-between border-b border-[#292720]/20 bg-[#eee9dc] px-4 py-3">
              <span className="font-mono text-[11px] font-bold">SQL CELL [1]</span>
              <span className="font-mono text-[10px] text-[#292720]/35">editable</span>
            </div>
            <textarea
              aria-label="SQL query"
              className="min-h-44 w-full resize-y bg-transparent p-5 font-mono text-sm leading-7 outline-none"
              onChange={(event) => demo.setQuery(event.target.value)}
              spellCheck={false}
              value={demo.query}
            />
            <div className="flex items-center justify-between border-t border-[#292720]/15 px-4 py-3">
              <span className="text-[11px] text-[#292720]/45">One statement at a time</span>
              <button
                className="border-2 border-[#292720] bg-[#ffcf70] px-5 py-2 text-xs font-bold shadow-[3px_3px_0_#292720] transition hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-[1px_1px_0_#292720] disabled:opacity-50"
                disabled={demo.running}
                onClick={() => void demo.runQuery()}
              >
                {demo.running ? "Evaluating…" : "Evaluate cell"}
              </button>
            </div>
          </article>

          <article className="mt-8 border-t-2 border-[#292720] pt-4">
            <div className="mb-3 flex items-baseline justify-between">
              <h2 className="font-serif text-xl font-bold">Observed output</h2>
              <span className="font-mono text-[10px] text-[#292720]/45">
                {demo.result.elapsedMs === null ? "not run" : `${demo.result.elapsedMs} milliseconds`}
              </span>
            </div>
            <pre
              className={`min-h-40 overflow-x-auto border border-[#292720]/15 bg-[#e9e4d7] p-5 font-mono text-xs leading-6 ${demo.result.error ? "text-[#b43f2c]" : "text-[#292720]/75"}`}
            >
              {demo.result.error || demo.result.output}
            </pre>
          </article>

          <div className="mt-8 grid gap-4 sm:grid-cols-3">
            {[
              ["INTEGER", "64-bit signed"],
              ["TEXT", "UTF-8 string"],
              ["BOOLEAN", "true / false"],
            ].map(([type, note]) => (
              <div className="border-t border-[#292720]/30 pt-3" key={type}>
                <p className="font-mono text-xs font-bold">{type}</p>
                <p className="mt-1 font-serif text-xs italic text-[#292720]/50">{note}</p>
              </div>
            ))}
          </div>
        </section>
      </div>
    </main>
  );
}
