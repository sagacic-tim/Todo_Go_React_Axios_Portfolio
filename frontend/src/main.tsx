// src/main.tsx

import React from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import { TasksProvider } from "./hooks/useTasks";
import "./styles/global.css";

const container = document.getElementById("app");
if (!container) throw new Error("`#app` not found");

createRoot(container).render(
  <React.StrictMode>
    <TasksProvider>
      <App />
    </TasksProvider>
  </React.StrictMode>
);
