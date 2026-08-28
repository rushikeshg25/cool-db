"use client";

import { useCallback, useEffect, useState } from "react";

export const demoQueries = [
  {
    label: "Create table",
    query:
      "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL, active BOOLEAN)",
  },
  {
    label: "Insert Ada",
    query: "INSERT INTO users VALUES (1, 'Ada Lovelace', true)",
  },
  { label: "Browse users", query: "SELECT * FROM users" },
  {
    label: "Update user",
    query: "UPDATE users SET active = false WHERE id = 1",
  },
] as const;

type ConnectionStatus = "checking" | "online" | "offline";

type QueryResult = {
  output: string;
  error: string;
  elapsedMs: number | null;
};

export function useDemoQuery() {
  const [query, setQuery] = useState<string>(demoQueries[2].query);
  const [connection, setConnection] =
    useState<ConnectionStatus>("checking");
  const [running, setRunning] = useState(false);
  const [result, setResult] = useState<QueryResult>({
    output: "Run a query to see results from your local CoolDB instance.",
    error: "",
    elapsedMs: null,
  });

  const checkConnection = useCallback(async () => {
    try {
      const response = await fetch("/api/health", { cache: "no-store" });
      setConnection(response.ok ? "online" : "offline");
    } catch {
      setConnection("offline");
    }
  }, []);

  useEffect(() => {
    let cancelled = false;
    fetch("/api/health", { cache: "no-store" })
      .then((response) => {
        if (!cancelled) setConnection(response.ok ? "online" : "offline");
      })
      .catch(() => {
        if (!cancelled) setConnection("offline");
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const runQuery = useCallback(
    async (nextQuery?: string) => {
      const statement = (nextQuery ?? query).trim();
      if (!statement || running) return;
      if (nextQuery) setQuery(nextQuery);

      setRunning(true);
      setResult({ output: "", error: "", elapsedMs: null });
      const startedAt = performance.now();
      try {
        const response = await fetch("/api/query", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ query: statement }),
        });
        const payload: {
          output?: string;
          error?: { message?: string };
        } = await response.json();
        const elapsedMs = Math.round(performance.now() - startedAt);
        if (!response.ok) {
          setResult({
            output: "",
            error: payload.error?.message ?? "Query failed.",
            elapsedMs,
          });
          if (response.status === 503) setConnection("offline");
          return;
        }
        setConnection("online");
        setResult({
          output: payload.output ?? "Query completed with no output.",
          error: "",
          elapsedMs,
        });
      } catch {
        setConnection("offline");
        setResult({
          output: "",
          error: "Could not reach the local CoolDB demo API.",
          elapsedMs: Math.round(performance.now() - startedAt),
        });
      } finally {
        setRunning(false);
      }
    },
    [query, running],
  );

  return {
    query,
    setQuery,
    connection,
    checkConnection,
    running,
    result,
    runQuery,
  };
}
