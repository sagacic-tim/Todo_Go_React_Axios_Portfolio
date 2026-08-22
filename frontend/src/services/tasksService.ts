// src/services/tasksService.ts
import { apiClient } from "./apiClient";
import type { Task } from "../types";

type TasksIndexResponse = { tasks: Task[] };

// If later your API wraps responses, you can switch these:
// type TaskResponse = { task: Task };

export type CreateTaskInput = {
  title: string;
  description: string;
  dueDate: string; // YYYY-MM-DD or ISO
};

export type UpdateTaskInput = Partial<CreateTaskInput> & {
  state?: Task["state"];
};

export async function listTasks(): Promise<Task[]> {
  const res = await apiClient.get<TasksIndexResponse>("/tasks");
  return res.data.tasks;
}

export async function createTask(input: CreateTaskInput): Promise<Task> {
  const res = await apiClient.post<Task>("/tasks", input);
  return res.data;
}

export async function updateTask(id: number, input: UpdateTaskInput): Promise<Task> {
  const res = await apiClient.patch<Task>(`/tasks/${id}`, input);
  return res.data;
}

export async function deleteTask(id: number): Promise<void> {
  await apiClient.delete(`/tasks/${id}`);
}
