// src/components/Modal.tsx
import React from "react";
import "./Modal.css";  // copy over your .open/.hidden/.dimmed rules

interface ModalProps {
  open: boolean;
  onClose: () => void;
  dimUnderneath?: boolean;
  children: React.ReactNode;
}

export const Modal: React.FC<ModalProps> = ({
  open,
  onClose,
  dimUnderneath = false,
  children,
}) => {
  return (
    <div
      className={[
        "lightbox-overlay",
        open ? "open" : "hidden",
        dimUnderneath ? "dimmed" : "",
      ].join(" ")}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div className="modal-content">{children}</div>
    </div>
  );
};

