// src/components/TaskEditorModal.tsx
import React, { useState, useEffect, useId, useRef } from 'react';
import ReactDOM from 'react-dom';
import styles from './TaskEditorModal.module.css';

export type TaskState = "" 
  | "scheduled"    // implicit on create 
  | "rescheduled"  // formerly “postponed”
  | "completed"
  | "dismissed";

export interface TaskData {
  title:       string;
  description: string;
  dueDate:     string;     // ISO YYYY-MM-DD
  state:       TaskState;
}

export interface TaskEditorModalProps {
  mode:           'Create Task' | 'Edit Task'
  initialTitle?:       string
  initialDescription?: string
  initialDueDate?:     string
  initialState?:       TaskState
  onClose:        () => void
  onSave:         (data: TaskData) => void
  onDelete?:      () => void       // ← allow parent to pass onDelete
}

export const TaskEditorModal: React.FC<TaskEditorModalProps> = ({
  mode,
  initialTitle       = '',
  initialDescription = '',
  initialDueDate     = '',
  initialState       = 'scheduled',
  onClose,
  onSave,
  onDelete
}) => {
  // 🍪 a single stable prefix for all IDs in this modal
  const uid = useId();  // form state
  const [title,       setTitle      ] = useState(initialTitle);
  const [description, setDescription] = useState(initialDescription);
  const [state,       setState]       = useState<TaskState>(initialState ?? "");
  const titleRef = useRef<HTMLTextAreaElement>(null);
  // whenever `title` changes, grow/shrink the textarea to fit its content:
  useEffect(() => {
    if (titleRef.current) {
     const ta = titleRef.current;
     ta.style.height = 'auto'; // reset any previous height
     ta.style.height = ta.scrollHeight + 'px';
    }
  }, [title]);

  // Initialize dueDate to a pure "YYYY-MM-DD" string:
  const normalizedDate = initialDueDate.includes("T")
    ? initialDueDate.slice(0,10)
    : initialDueDate;
  const [dueDate, setDueDate] = useState(normalizedDate);

  // slide‐in animation
  const [isMounted, setIsMounted] = useState(false);
  useEffect(() => {
    const t = window.setTimeout(() => setIsMounted(true), 0);
    return () => window.clearTimeout(t);
  }, []);

  const isEdit = mode === 'Edit Task';
  // In Edit mode, only allow dueDate changes if “rescheduled”
  const dueDateDisabled = isEdit && state !== 'rescheduled';

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave({ title, description, dueDate, state });
    onClose();
  };

  const editStates: TaskState[] = ['rescheduled','completed','dismissed'];

  // unique IDs for accessibility
  const idTitle       = `${uid}-task-title`;
  const idDescription = `${uid}-task-description`;
  const idDueDate     = `${uid}-task-due-date`;

  return ReactDOM.createPortal(
    <div className={styles.backdrop} onClick={onClose}>
      <div
        className={`${styles.editorModal} ${isMounted ? styles.open : ''}`}
        onClick={e => e.stopPropagation()}
      >
        <button className={styles.closeBtn} onClick={onClose}>✕</button>
        <h2 className={styles.header}>
          <span>{mode}</span>
          <span className={styles.icon}>✍🏼</span>
        </h2>

        <form className={styles.taskForm} onSubmit={handleSubmit}>
          {/* Title */}
          <div className={styles.titleBlock}>
            <label htmlFor={idTitle}>Task Name:</label>
            <textarea
              id={idTitle}
              ref={titleRef}
              className={styles.taskTitle}
              value={title}
              onChange={e => setTitle(e.target.value)}
              required
            />
          </div>

          {/* Description */}
          <div className={styles.descriptionBlock}>
            <label htmlFor={idDescription}>Description:</label>
            <textarea
              id={idDescription}
              className={styles.taskDescription}
              value={description}
              onChange={e => setDescription(e.target.value)}
              required
            />
          </div>

          {/* Due Date */}
          <div className={styles.dueDateBlock}>
            <label htmlFor={idDueDate}>
              {isEdit
                ? state === 'rescheduled'
                  ? 'New Due Date:'
                  : 'Due Date (locked)'
                : 'When is Task Due:'}
            </label>
            <input
              id={idDueDate}
              className={styles.taskDueDate}
              type="date"
              value={dueDate}
              onChange={e => setDueDate(e.target.value)}
              disabled={dueDateDisabled}
              required
            />
          </div>

          {/* State radios (only in Edit mode) */}
          {isEdit && (
            <div className={styles.stateOfTask}>
              <label>State:</label>
              <fieldset className={styles.radioGroup}>
                {editStates.map(s => {
                  const isChecked = state === s;
                  return (
                    <label
                      key={s}
                      // toggle‐on/toggle‐off by clicking the label
                      className={isChecked ? styles.checkedLabel : ""}
                      onClick={() => setState(isChecked ? "" : s)}
                    >
                      <input
                        type="radio"
                        name={`${uid}-task-state`}
                        value={s}
                        checked={isChecked}
                        // prevent the browser’s default toggle,
                        // we'll drive it from the label onClick
                        onChange={() => {}}
                      />
                      {s.charAt(0).toUpperCase() + s.slice(1)}
                    </label>
                  );
                })}
              </fieldset>
            </div>
          )}

          {/* Actions */}
          <div className={styles.actions}>
            <button type="submit"
              className={styles.saveBtn}
            >
              Save Task
            </button>
            <button
              type="button"
              className={styles.cancelBtn}
              onClick={onClose}
            >
              Cancel
            </button>
            {mode === 'Edit Task' && onDelete && (
            <button
              type="button"
              className={styles.deleteBtn}
              onClick={onDelete}
            >
              Delete Task
            </button>
          )}
          </div>
        </form>
      </div>
    </div>,
    document.body
  );
};
