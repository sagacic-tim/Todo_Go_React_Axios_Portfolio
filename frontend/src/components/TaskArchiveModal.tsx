
import React, { useState, useEffect } from 'react'
import ReactDOM from 'react-dom'
import { Task } from '../types'
import styles from './TaskArchiveModal.module.css'

interface Props {
  tasks: Task[];
  onClose: () => void;
  onEdit: (t: Task) => void;
  onCreateNew?: () => void;
  initialFilter?: "all"|"completed"|"rescheduled"|"dismissed"|"delinquent";
}

export const TaskArchiveModal: React.FC<Props> = ({
  tasks,
  onClose,
  onEdit,
  onCreateNew,
  initialFilter = "all",      // ← default to “all”
  }) => {

  const [isMounted, setIsMounted] = useState(false)
  useEffect(() => {
    const id = window.setTimeout(() => setIsMounted(true), 0)
    return () => window.clearTimeout(id)
  }, [])
  const todayIso = new Date().toISOString().slice(0,10)
  const [stateFilter, setStateFilter] = useState(initialFilter);
  const [from, setFrom] = useState<string>("");
  const [to,   setTo]   = useState<string>("");

const filtered = tasks
  .filter(t => {
    const iso = t.dueDate.slice(0,10);

    // 1) state logic
    let stateOK = true;
    if (stateFilter !== 'all') {
      switch (stateFilter) {
        case 'completed':   stateOK =  t.state === 'completed'; break;
        case 'rescheduled': stateOK =  t.wasRescheduled;        break;
        case 'dismissed':   stateOK =  t.wasDismissed;          break;
        case 'delinquent':  
          stateOK =  iso < todayIso && t.state !== 'completed';
          break;
      }
    }

    // 2) date-range logic
    const rangeOK = (!from || iso >= from) && (!to || iso <= to);

    return stateOK && rangeOK;
  })
  .sort((a, b) => a.dueDate.localeCompare(b.dueDate));


  return ReactDOM.createPortal(
    <div className={styles.backdrop} onClick={onClose}>
      <div
        className={[
          styles.archiveModal,
          isMounted && styles.open  // <-- this picks up `.archiveModal.open`
        ].filter(Boolean).join(' ')}
        onClick={e => e.stopPropagation()}
      >
        <button className={styles.closeBtn} onClick={onClose}>✕</button>
        <header className={styles.header}>
          <h2>Task Archive</h2>
          <button onClick={onClose}>✕</button>
        </header>

        <div className={styles.controls}>
          <div className={styles.statuses}>
            <label htmlFor="statusSelect">Select Status:</label>
            <select id="statusSelect" value={stateFilter} onChange={e => setStateFilter(e.target.value as any)}>
              <option value="all">All</option>
              <option value="completed">Completed</option>
              <option value="rescheduled">Rescheduled</option>
              <option value="dismissed">Dismissed</option>
              <option value="delinquent">Delinquent</option>
            </select>
          </div>
          <div className={styles.dateRange}>
            <label htmlFor="dateRangeBlock">Select Date Range:</label>
            <div id="dateRangeBlock" className={styles.dateRangeSelection}>
              <label>
                From
                <input type="date" value={from} onChange={e => setFrom(e.target.value)} />
              </label>
              <label>
                To
                <input type="date" value={to}   onChange={e => setTo(e.target.value)} />
              </label>
            </div>
          </div>
        </div>
        <hr className={styles.selectionsEnd} />

        {onCreateNew && (
          <button
            className={styles.createBtn}
            onClick={onCreateNew}
          >
            ＋ Create New Task
          </button>
        )}

        <ul className={styles.taskList}>
          {filtered.map(t => (
            <li key={t.id}>
              <strong>{new Date(t.dueDate).toLocaleDateString()}</strong> –
              <button
                className={styles.taskLink}
                onClick={() => onEdit(t)}
              >
                {t.title}
              </button>
            </li>
          ))}
        </ul>
      </div>
    </div>,
    document.body
  )
}
