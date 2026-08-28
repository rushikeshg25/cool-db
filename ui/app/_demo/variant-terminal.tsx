"use client";

import { demoQueries, useDemoQuery } from "./use-demo-query";

export const variantName = "Terminal wall";

export function VariantTerminal() {
  const demo = useDemoQuery();

  return (
    <main className="min-h-screen overflow-hidden bg-[#050806] font-mono text-[#b7f7c8] selection:bg-[#b7f7c8] selection:text-[#050806]">
      <div className="pointer-events-none fixed inset-0 bg-[linear-gradient(rgba(183,247,200,.025)_1px,transparent_1px)] bg-[size:100%_4px]" />
      <header className="relative flex items-center justify-between border-b border-[#b7f7c8]/15 px-4 py-3 sm:px-7">
        <div className="flex items-center gap-4">
          <span className="text-sm font-bold tracking-[.2em]">COOLDB_OS</span>
          <span className="hidden text-[10px] text-[#b7f7c8]/35 sm:block">SESSION 0X3040</span>
        </div>
        <div className="flex items-center gap-2 text-[10px]">
          <span className={demo.connection === "online" ? "text-[#74ff99]" : "text-[#ff7b72]"}>●</span>
          API_{demo.connection.toUpperCase()}
        </div>
      </header>

      <div className="relative grid min-h-[calc(100vh-45px)] xl:grid-cols-[1fr_310px]">
        <section className="flex min-w-0 flex-col p-4 sm:p-7 lg:p-10">
          <div className="mb-8">
            <p className="text-[10px] tracking-[.28em] text-[#b7f7c8]/35">LOCAL DATABASE INTERFACE</p>
            <h1 className="mt-3 text-2xl font-bold tracking-[-.04em] sm:text-4xl">
              DATA TERMINAL <span className="animate-pulse">_</span>
            </h1>
          </div>

          <div className="mb-5 flex flex-wrap gap-x-5 gap-y-2 border-y border-[#b7f7c8]/10 py-3 text-[10px] text-[#b7f7c8]/45">
            {demoQueries.map((example, index) => (
              <button
                className="transition hover:text-[#b7f7c8]"
                key={example.label}
                onClick={() => demo.setQuery(example.query)}
              >
                [F{index + 1}] {example.label.toUpperCase()}
              </button>
            ))}
          </div>

          <div className="border border-[#b7f7c8]/20 bg-[#071009] shadow-[0_0_70px_rgba(75,255,125,.04)]">
            <div className="flex items-center justify-between border-b border-[#b7f7c8]/15 px-4 py-2 text-[10px] text-[#b7f7c8]/40">
              <span>INPUT_BUFFER</span>
              <span>CTRL+ENTER / EXECUTE</span>
            </div>
            <div className="flex p-4 sm:p-6">
              <span className="mr-3 select-none text-[#74ff99]">cooldb&gt;</span>
              <textarea
                aria-label="SQL query"
                className="min-h-40 flex-1 resize-y bg-transparent text-sm leading-6 text-[#d5ffdf] caret-[#74ff99] outline-none"
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
            <div className="flex justify-end border-t border-[#b7f7c8]/15 p-3">
              <button
                className="border border-[#74ff99] bg-[#74ff99]/10 px-5 py-2 text-[11px] font-bold tracking-wider text-[#9dffb5] transition hover:bg-[#74ff99] hover:text-[#050806] disabled:animate-pulse disabled:opacity-50"
                disabled={demo.running}
                onClick={() => void demo.runQuery()}
              >
                {demo.running ? "PROCESSING..." : "EXECUTE ↵"}
              </button>
            </div>
          </div>

          <div className="mt-5 flex-1 border border-[#b7f7c8]/15 bg-black/30">
            <div className="flex items-center justify-between border-b border-[#b7f7c8]/10 px-4 py-2 text-[10px] text-[#b7f7c8]/35">
              <span>STDOUT</span>
              <span>{demo.result.elapsedMs === null ? "IDLE" : `DONE / ${demo.result.elapsedMs}MS`}</span>
            </div>
            <pre
              className={`min-h-52 overflow-x-auto p-5 text-xs leading-6 sm:p-6 ${demo.result.error ? "text-[#ff7b72]" : "text-[#b7f7c8]/70"}`}
            >
              {demo.result.error ? `ERROR: ${demo.result.error}` : demo.result.output}
            </pre>
          </div>
        </section>

        <aside className="hidden border-l border-[#b7f7c8]/15 p-6 xl:block">
          <Section title="SYSTEM">
            <TerminalRow label="ENGINE" value="COOLDB/0.1" />
            <TerminalRow label="MODE" value="LOCAL_DEMO" />
            <TerminalRow label="STORAGE" value="SNAPSHOT" />
            <TerminalRow label="ENCODING" value="UTF-8" />
          </Section>

          <Section title="SCHEMA_MAP">
            <div className="text-[11px] leading-6 text-[#b7f7c8]/55">
              <p>└─ default.cooldb</p>
              <p className="pl-4">└─ users</p>
              <p className="pl-8">├─ id : INTEGER</p>
              <p className="pl-8">├─ name : TEXT</p>
              <p className="pl-8">└─ active : BOOLEAN</p>
            </div>
          </Section>

          <Section title="SUPPORTED_OPS">
            <div className="grid grid-cols-2 gap-2 text-[10px]">
              {["CREATE", "DROP", "INSERT", "SELECT", "UPDATE", "DELETE"].map((operation) => (
                <span className="border border-[#b7f7c8]/10 px-2 py-2 text-center text-[#b7f7c8]/45" key={operation}>
                  {operation}
                </span>
              ))}
            </div>
          </Section>

          <div className="mt-8 border border-[#ffcf70]/25 p-4 text-[10px] leading-5 text-[#ffcf70]/60">
            WARNING: PROTOTYPE INTERFACE. QUERIES MAY MODIFY THE CONNECTED SCRATCH DATABASE.
          </div>
        </aside>
      </div>
    </main>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="mb-8">
      <p className="mb-4 border-b border-[#b7f7c8]/10 pb-2 text-[10px] tracking-[.2em] text-[#b7f7c8]/35">
        /{title}
      </p>
      {children}
    </section>
  );
}

function TerminalRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="mb-2 flex justify-between text-[10px]">
      <span className="text-[#b7f7c8]/30">{label}</span>
      <span className="text-[#b7f7c8]/65">{value}</span>
    </div>
  );
}
