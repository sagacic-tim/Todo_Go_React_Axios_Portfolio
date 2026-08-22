// src/types.ts
export type TaskState =
  | "scheduled"
  | "rescheduled"
  | "completed"
  | "dismissed";

export interface Task {
  id: number;
  title: string;
  description?: string;
  dueDate: string;       // ISO “YYYY-MM-DD”
  state: TaskState;      // must match the enum values
  wasRescheduled: boolean;
  wasDismissed:   boolean;
}
