// src/hooks/useTasks.tsx
import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { Task } from "../types";
import type { CreateTaskInput, UpdateTaskInput } from "../services/tasksService";
import {
  listTasks,
  createTask,
  updateTask,
  deleteTask as deleteTaskApi,
} from "../services/tasksService";
import { toMessage } from "../services/apiClient";

type UseTasksValue = {
  tasks: Task[];
  isLoading: boolean;
  error: string | null;

  refresh: () => Promise<void>;
  save: (opts: { id?: number; data: CreateTaskInput | UpdateTaskInput }) => Promise<void>;
  remove: (id: number) => Promise<void>;

  incompleteTasks: Task[];
};

const TasksContext = createContext<UseTasksValue | null>(null);

function useTasksInternal(): UseTasksValue {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Prevent state updates after unmount (HMR/route changes)
  const isMountedRef = useRef(true);
  useEffect(() => {
    isMountedRef.current = true;
    return () => {
      isMountedRef.current = false;
    };
  }, []);

  // Avoid flicker if refresh overlaps (e.g. save -> refresh while another refresh runs)
  const inflight = useRef(0);

  const refresh = useCallback(async () => {
    inflight.current += 1;

    if (isMountedRef.current) {
      setIsLoading(true);
      setError(null);
    }

    try {
      const t = await listTasks();
      if (isMountedRef.current) setTasks(t);
    } catch (err) {
      if (isMountedRef.current) setError(toMessage(err));
    } finally {
      inflight.current -= 1;
      if (inflight.current <= 0) inflight.current = 0;
      if (isMountedRef.current && inflight.current === 0) setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const save = useCallback(
    async (opts: { id?: number; data: CreateTaskInput | UpdateTaskInput }) => {
      if (isMountedRef.current) setError(null);

      try {
        if (typeof opts.id === "number") {
          await updateTask(opts.id, opts.data as UpdateTaskInput);
        } else {
          await createTask(opts.data as CreateTaskInput);
        }
        await refresh();
      } catch (err) {
        if (isMountedRef.current) setError(toMessage(err));
        throw err;
      }
    },
    [refresh]
  );

  const remove = useCallback(
    async (id: number) => {
      if (isMountedRef.current) setError(null);

      try {
        await deleteTaskApi(id);
        await refresh();
      } catch (err) {
        if (isMountedRef.current) setError(toMessage(err));
        throw err;
      }
    },
    [refresh]
  );

  const incompleteTasks = useMemo(
    () => tasks.filter((t) => t.state !== "completed"),
    [tasks]
  );

  return { tasks, isLoading, error, refresh, save, remove, incompleteTasks };
}

export function TasksProvider({ children }: { children: React.ReactNode }) {
  const value = useTasksInternal();
  return <TasksContext.Provider value={value}>{children}</TasksContext.Provider>;
}

export function useTasks(): UseTasksValue {
  const ctx = useContext(TasksContext);
  if (!ctx) {
    throw new Error("useTasks must be used within <TasksProvider>");
  }
  return ctx;
}
