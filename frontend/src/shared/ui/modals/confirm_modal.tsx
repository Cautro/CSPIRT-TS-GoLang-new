import { useEffect } from "react";
import {createPortal} from "react-dom";

interface ConfirmModalProps {
    title?: string;
    content: string;
    buttonContent: string;
    isDanger?: boolean;

    isOpen: boolean;
    onClose: () => void;
    onConfirm: () => Promise<void>;
}

export function ConfirmModal({title = "Подтвердите действие", isOpen, onClose, content, buttonContent, onConfirm, isDanger = false,}: ConfirmModalProps) {
    useEffect(() => {
        if (!isOpen) {
            return;
        }

        function handleEscape(event: KeyboardEvent) {
            if (event.key === "Escape") {
                onClose();
            }
        }

        document.addEventListener("keydown", handleEscape);
        document.body.style.overflow = "hidden";

        return () => {
            document.removeEventListener("keydown", handleEscape);
            document.body.style.overflow = "";
        };
    }, [isOpen, onClose]);

    if (!isOpen) {
        return null;
    }

    return createPortal(
        <div className="modal-backdrop" onMouseDown={onClose}>
            <section
                className="modal modal--confirm"
                role="dialog"
                aria-modal="true"
                aria-labelledby="confirm-modal-title"
                onMouseDown={(event) => event.stopPropagation()}
            >
                <div className="modal__header">
                    <div>
                        <h2 className="modal__title" id="confirm-modal-title">
                            {title}
                        </h2>

                        <p className="modal__description">
                            {content}
                        </p>
                    </div>

                    <button
                        className="modal__close"
                        type="button"
                        onClick={onClose}
                        aria-label="Закрыть окно подтверждения"
                    >
                        ×
                    </button>
                </div>

                <div className="modal__footer">
                    <button
                        className="btn btn--secondary"
                        type="button"
                        onClick={onClose}
                    >
                        Отмена
                    </button>

                    <button
                        className={isDanger ? "btn btn--danger" : "btn btn--primary"}
                        type="button"
                        onClick={() => void onConfirm()}
                    >
                        {buttonContent}
                    </button>
                </div>
            </section>
        </div>,
        document.body
    );
}