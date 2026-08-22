// src/views/TaskListView.tsx
import React from "react";
import { useTasks } from "../hooks/useTasks";
import type { Task } from "../types";
import styles from "./TaskListView.module.css";

// Today's date as "YYYY-MM-DD" (local time, no timezone shift)
const todayIso = new Date().toLocaleDateString("en-CA"); // "YYYY-MM-DD"

// Format "YYYY-MM-DD" → "April 3, 2026"
function formatDate(iso: string): string {
  const [year, month, day] = iso.slice(0, 10).split("-").map(Number);
  return new Date(year, month - 1, day).toLocaleDateString(undefined, {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

// A task is delinquent when it is still open (scheduled or rescheduled)
// but its due date has already passed — matching the calendar's logic.
function effectiveState(task: Task): string {
  if (
    (task.state === "scheduled" || task.state === "rescheduled") &&
    task.dueDate.slice(0, 10) < todayIso
  ) {
    return "delinquent";
  }
  return task.state;
}

// Capitalise first letter for display
function formatState(state: string): string {
  return state.charAt(0).toUpperCase() + state.slice(1);
}

export default function TaskListView() {
  const { tasks, isLoading, error } = useTasks();

  // Delinquent tasks first, then newest due date first within each group
  const sorted = [...tasks].sort((a, b) => {
    const da = effectiveState(a) === "delinquent" ? 0 : 1;
    const db = effectiveState(b) === "delinquent" ? 0 : 1;
    if (da !== db) return da - db;
    return b.dueDate.localeCompare(a.dueDate);
  });

  return (
    <main className={styles.page}>
      <h1 className={styles.heading}>All Tasks</h1>

      {isLoading && <p className={styles.status}>Loading…</p>}
      {error   && <p className={styles.statusError}>{error}</p>}

      {!isLoading && sorted.length === 0 && (
        <p className={styles.status}>No tasks yet.</p>
      )}

      <div className={styles.grid}>
        {sorted.map((task) => {
          const display = effectiveState(task);
          return (
            <div key={task.id} className={`${styles.card} ${styles[display]}`}>
              <p className={styles.meta}>
                <span className={styles.dueDate}>{formatDate(task.dueDate)}</span>
                <span className={styles.separator}>·</span>
                <span className={styles.title}>{task.title}</span>
                <span className={styles.separator}>·</span>
                <span className={styles.state}>{formatState(display)}</span>
              </p>
              <p className={styles.description}>
                {task.description || <em>No description.</em>}
              </p>
            </div>
          );
        })}
      </div>
    </main>
  );
}
