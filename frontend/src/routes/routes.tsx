// src/routes/routes.tsx
import { Route, Routes } from "react-router-dom";
import { CalendarGrid } from "../components/CalendarGrid";
import TaskListView from "../views/TaskListView";

export default function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<CalendarGrid />} />
      <Route path="/tasks" element={<TaskListView />} />
      {/* later you can add routes for “new” and “edit” pages */}
    </Routes>
  );
}
