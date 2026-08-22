// src/components/CalendarGrid.tsx
import React, { useState, useMemo, useEffect } from "react";
import ReactDOM from "react-dom";
import styles from "./CalendarGrid.module.css";
import { TaskListModal } from "./TaskListModal";
import { TaskEditorModal, TaskData, TaskState } from "./TaskEditorModal";
import { TaskArchiveModal } from "./TaskArchiveModal";
import { Task } from "../types";

import { useTasks } from "../hooks/useTasks";

interface DayCell {
  date: Date;
  inCurrentMonth: boolean;
}

export const CalendarGrid: React.FC = () => {
  const today = new Date();
  const todayIso = today.toISOString().slice(0, 10);

  // Tasks store (shared across routes via provider)
  const { tasks, incompleteTasks, isLoading, error, save, remove } = useTasks();

  // how many incomplete tasks are already past their due date?
  const totalDelinquent = incompleteTasks.filter(
    (t) => t.dueDate.slice(0, 10) < todayIso
  ).length;

  // pick your label
  let delinquentLabel = "";
  if (totalDelinquent === 1) {
    delinquentLabel = `[1] task delinquent`;
  } else if (totalDelinquent > 1) {
    delinquentLabel = `[${totalDelinquent}] tasks delinquent`;
  }

  // 2) Build a 42-cell month grid
  const [year, setYear] = useState(today.getFullYear());
  const [month, setMonth] = useState(today.getMonth());

  const days: DayCell[] = useMemo(() => {
    const firstOfMonth = new Date(year, month, 1);
    const startDow = firstOfMonth.getDay(); // 0–6
    const daysInMonth = new Date(year, month + 1, 0).getDate();
    const daysInPrev = new Date(year, month, 0).getDate();
    const cells: DayCell[] = [];

    const mk = (d: Date): DayCell => ({
      date: d,
      inCurrentMonth: d.getMonth() === month,
    });

    // leading prev-month days
    for (let i = startDow - 1; i >= 0; i--) {
      cells.push(mk(new Date(year, month - 1, daysInPrev - i)));
    }
    // current month
    for (let d = 1; d <= daysInMonth; d++) {
      cells.push(mk(new Date(year, month, d)));
    }
    // trailing next-month days
    while (cells.length < 42) {
      const idx = cells.length - (startDow + daysInMonth);
      cells.push(mk(new Date(year, month + 1, idx + 1)));
    }
    return cells;
  }, [year, month]);

  // 3) Calendar navigation
  const prevYear = () => setYear((y) => y - 1);
  const nextYear = () => setYear((y) => y + 1);

  const prevMonth = () => {
    setMonth((m) => (m === 0 ? 11 : m - 1));
    if (month === 0) setYear((y) => y - 1);
  };

  const nextMonth = () => {
    setMonth((m) => (m === 11 ? 0 : m + 1));
    if (month === 11) setYear((y) => y + 1);
  };

  const monthNames = [
    "January",
    "February",
    "March",
    "April",
    "May",
    "June",
    "July",
    "August",
    "September",
    "October",
    "November",
    "December",
  ];

  const weekDayNames = [
    "Sunday",
    "Monday",
    "Tuesday",
    "Wednesday",
    "Thursday",
    "Friday",
    "Saturday",
  ];

  // 4) Modal & editor state
  const [selectedDate, setSelectedDate] = useState<Date | null>(null);
  const [showList, setShowList] = useState(false);
  const [showEditor, setShowEditor] = useState(false);
  const [showArchive, setShowArchive] = useState(false);
  const [editorTask, setEditorTask] = useState<Task | null>(null);

  const selectedIso = selectedDate?.toISOString().slice(0, 10) ?? "";

  const handleCreateNew = () => {
    setEditorTask(null);
    setShowEditor(true);
  };

  const handleEdit = (task: Task) => {
    setEditorTask(task);
    setShowEditor(true);
  };

  const handleArchiveEdit = (task: Task) => {
    setShowArchive(false);
    handleEdit(task);
  };

  const [archiveFilter, setArchiveFilter] = useState<
    "all" | "completed" | "rescheduled" | "dismissed" | "delinquent"
  >("all");

  // 5) Backdrop for modals
  const [backdropOpen, setBackdropOpen] = useState(false);

  useEffect(() => {
    setBackdropOpen(showList || showEditor || showArchive);
  }, [showList, showEditor, showArchive]);

  const openList = (d: Date) => {
    setSelectedDate(d);
    setShowList(true);
  };

  const closeList = () => {
    setShowList(false);
    setEditorTask(null);
    setSelectedDate(null);
  };

  const backdropPortal =
    showList || showEditor || showArchive
      ? ReactDOM.createPortal(
          <div
            className={`${styles.backdrop} ${backdropOpen ? styles.open : ""}`}
            onClick={() => {
              if (showEditor) setShowEditor(false);
              else closeList();
            }}
          />,
          document.body
        )
      : null;

  // 6) Save / Delete handlers (via hook)
  const saveTask = async (data: TaskData) => {
    try {
      if (editorTask) {
        await save({ id: editorTask.id, data });
      } else {
        await save({ data });
      }
      // If you prefer to keep the editor open on error, move setShowEditor(false)
      // into the try block and let errors throw.
    } finally {
      setShowEditor(false);
    }
  };

  const deleteTask = async () => {
    if (!editorTask || !confirm("Delete this task?")) return;

    try {
      await remove(editorTask.id);
    } finally {
      setShowEditor(false);
    }
  };

  // Render
  return (
    <>
      <div className={styles.calendarWrapper}>
        <div className={styles.calendarTitle}>
          Groovey Task Manager

          {totalDelinquent > 0 && (
            <button
              className={styles.delinquentAlert}
              onClick={() => {
                setArchiveFilter("delinquent");
                setShowArchive(true);
              }}
              title="View All Delinquent Tasks"
            >
              <span className={styles.cautionSign}>⚠️</span> {delinquentLabel}
            </button>
          )}
        </div>

        <div className={styles.calendarBodyWrapper}>
          <div className={styles.navBar}>
            <button
              onClick={() => {
                setArchiveFilter("all");
                setShowArchive(true);
              }}
            >
              View Archive
            </button>

            <button onClick={prevYear}>&laquo; Year</button>
            <button onClick={prevMonth}>&lsaquo; Month</button>

            <span className={styles.title}>
              {monthNames[month]} {year}
            </span>

            <button onClick={nextMonth}>Month &rsaquo;</button>
            <button onClick={nextYear}>Year &raquo;</button>

            <button
              onClick={() => {
                setYear(today.getFullYear());
                setMonth(today.getMonth());
              }}
            >
              Current Month
            </button>

            {isLoading && <span style={{ marginLeft: 8 }}>Loading…</span>}
          </div>

          {/* Optional: surface API errors non-intrusively */}
          {error && (
            <div style={{ marginTop: 8, fontSize: 14 }}>
              <strong>Error:</strong> {error}
            </div>
          )}
        </div>

        <ul className={styles.weekdays}>
          {weekDayNames.map((wd) => (
            <li key={wd} className={styles.weekdayItem}>
              <abbr title={wd}>{wd}</abbr>
            </li>
          ))}
        </ul>

        <ol className={styles.dayGrid}>
          {days.map(({ date, inCurrentMonth }, i) => {
            const cellIso = date.toISOString().slice(0, 10);
            const isTodayCell = cellIso === todayIso;

            // tasks due on this specific day
            const dueTasks = incompleteTasks.filter(
              (t) => t.dueDate.slice(0, 10) === cellIso
            );
            const dueCount = dueTasks.length;

            // only show delinquent badge on today’s cell
            const delinquentCount = isTodayCell
              ? incompleteTasks.filter((t) => t.dueDate.slice(0, 10) < todayIso)
                  .length
              : 0;

            return (
              <li
                key={i}
                className={[
                  styles.dayCell,
                  !inCurrentMonth && styles.faded,
                  isTodayCell && styles.currentDay,
                ]
                  .filter(Boolean)
                  .join(" ")}
                onClick={() => openList(date)}
              >
                <div className={styles.calendarDay}>
                  <span className={styles.dayNo}>{date.getDate()}</span>

                  {dueCount > 0 && (
                    <p className={styles.dueLabel}>
                      {dueCount === 1
                        ? "1 task due today"
                        : `${dueCount} tasks due today`}
                    </p>
                  )}

                  {delinquentCount > 0 && (
                    <p
                      className={`${styles.todoCount} ${styles.pastDue}`}
                      data-count={delinquentCount}
                    >
                      Delinquent
                    </p>
                  )}
                </div>
              </li>
            );
          })}
        </ol>
      </div>

      {backdropPortal}

      {/* Task List Modal */}
      {showList && selectedDate && (
        <TaskListModal
          dateLabel={selectedDate.toLocaleDateString(undefined, {
            year: "numeric",
            month: "long",
            day: "numeric",
          })}
          selectedIso={selectedDate.toISOString().slice(0, 10)}
          tasks={tasks}
          onClose={closeList}
          onCreateNew={handleCreateNew}
          onEdit={handleEdit}
        />
      )}

      {/* Task Editor Modal */}
      {showEditor && (
        <TaskEditorModal
          mode={editorTask ? "Edit Task" : "Create Task"}
          initialTitle={editorTask?.title}
          initialDescription={editorTask?.description}
          initialDueDate={editorTask ? editorTask.dueDate.slice(0, 10) : selectedIso}
          initialState={(editorTask?.state as TaskState) ?? "scheduled"}
          onClose={() => setShowEditor(false)}
          onSave={saveTask}
          onDelete={deleteTask}
        />
      )}

      {/* Task Archive Modal */}
      {showArchive && (
        <TaskArchiveModal
          tasks={tasks}
          onClose={() => setShowArchive(false)}
          onEdit={handleArchiveEdit}
          onCreateNew={handleCreateNew}
          initialFilter={archiveFilter}
        />
      )}
    </>
  );
};
