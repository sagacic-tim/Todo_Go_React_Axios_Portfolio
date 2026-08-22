// src/components/TaskListModal.tsx
import React, { useState, useEffect } from "react";
import ReactDOM from "react-dom";
import styles from "./TaskListModal.module.css";
import { Task } from "../types";

interface Props {
  dateLabel:   string;   // e.g. “July 4, 2025”
  selectedIso: string;   // e.g. “2025-07-04”
  tasks:       Task[];   // _all_ non-completed tasks
  onClose:     () => void;
  onCreateNew: () => void;
  onEdit:      (t: Task) => void;
}

export const TaskListModal: React.FC<Props> = ({
  dateLabel,
  selectedIso,
  tasks,
  onClose,
  onCreateNew,
  onEdit
}) => {
  // slide-in animation
  const [isMounted, setIsMounted] = useState(false);
  useEffect(() => {
    const id = window.setTimeout(() => setIsMounted(true), 0);
    return () => {
      window.clearTimeout(id);
      setIsMounted(false);
    };
  }, []);


  // figure out header text
  const todayIso = new Date().toISOString().slice(0,10);
  const isToday  = selectedIso === todayIso;
  // turn "2025-07-04" → ["2025","07","04"]
  const [Y, M, D] = selectedIso.split("-");
  const numeric = `${M}/${D}/${Y}`;  // "07/04/2025"

  const dueHeader = isToday ? "Due Today" : `Due ${numeric}`;
  const noDueText = isToday 
    ? "No tasks due today." 
    : `No tasks due ${numeric}.`;

  // split your tasks
  const tasksDue = tasks.filter(
    t => t.dueDate.slice(0,10) === selectedIso &&
    t.state !== "completed"
  );

  const tasksDelinquent = tasks.filter(
    t =>
      t.dueDate.slice(0,10) < selectedIso &&
      t.state !== "completed"
  );

  return ReactDOM.createPortal(
    <div
      className={[
        styles.listModal,
        isMounted && styles.open
      ].filter(Boolean).join(" ")}
      onClick={e => e.stopPropagation()}
    >
      <button className={styles.closeBtn} onClick={onClose}>✕</button>
      <h2 className={styles.taskListTitle}>Tasks for {dateLabel}</h2>
      <button className={styles.createBtn} onClick={onCreateNew}>
        ＋ Create New Task
      </button>

      {/* — Due section — */}
      <section className={styles.dueTodayContainer}>
        <h3>{dueHeader}</h3>
        <div className={styles.tasksContainer}>
          {tasksDue.length === 0
            ? <p>{noDueText}</p>
            : <ul className={styles.taskList}>
                {tasksDue.map(t => (
                  <li key={t.id}>
                    <button
                      className={styles.taskLink}
                      onClick={() => onEdit(t)}
                    >
                      {t.title}
                    </button>
                  </li>
                ))}
              </ul>
          }
        </div>
      </section>

      {/* — Delinquent section — */}
      <section className={styles.delinquentContainer}>
        <h3>Delinquent</h3>
        <div className={styles.tasksContainer}>
          {tasksDelinquent.length === 0
            ? <p>No delinquent tasks.</p>
            : <ul className={styles.taskList}>
                {tasksDelinquent.map(t => (
                  <li key={t.id}>
                    <button
                      className={styles.taskLink}
                      onClick={() => onEdit(t)}
                    >
                      {t.title}
                    </button>
                  </li>
                ))}
              </ul>}
        </div>
      </section>
    </div>,
    document.body
  );
};
