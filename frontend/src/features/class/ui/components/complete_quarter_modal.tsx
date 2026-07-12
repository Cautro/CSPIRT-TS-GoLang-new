import {useEffect, useState} from "react";
import {useCompleteQuarter} from "../../hooks/use_complete_quarter.ts";
import type {ClassType} from "../../../../shared/entities/class/types/class_types.ts";

interface CompleteQuarterModalProps {
    isOpen: boolean;
    onClose: () => void;
    parallelId: number;
}

export function CompleteQuarterModal({isOpen, onClose, parallelId}: CompleteQuarterModalProps) {

    const [completed, setCompleted] = useState(false);

    const {mutateAsync, isPending, isError, error, data} = useCompleteQuarter();

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

    useEffect(() => {
        if (!isOpen) {
            setCompleted(false);
        }
    }, [isOpen]);

    if (!isOpen) {
        return null;
    }

    async function handleCompleteQuarter() {
        try {
            await mutateAsync({id: parallelId});
            setCompleted(true);
        } catch (e) {
            console.error(e);
        }
    }

    function renderClassPlace(
        title: string,
        classItem: ClassType | null
    ) {
        return (
            <div className="quarter-place-card">
                <div className="quarter-place-card__header">
                    <h3>{title}</h3>
                </div>

                {!classItem ? (
                    <p className="quarter-place-card__empty">
                        Нет данных
                    </p>
                ) : (
                    <div className="quarter-place-card__body">
                        <p>
                            <strong>Класс:</strong> {classItem.Name}
                        </p>

                        <p>
                            <strong>Общий рейтинг:</strong>{" "}
                            {classItem.ClassTotalRating +
                                classItem.UserTotalRating}
                        </p>

                        <p>
                            <strong>Ученики:</strong>{" "}
                            {classItem.Members.length}
                        </p>
                    </div>
                )}
            </div>
        );
    }

    return (
        <div
            className="modal-backdrop"
            onMouseDown={() => {
                onClose();
            }}
        >
            <section
                className="modal modal--wide"
                role="dialog"
                aria-modal="true"
                aria-labelledby="complete-quarter-modal-title"
                onMouseDown={(event) => event.stopPropagation()}
            >

                <div className="modal__header">
                    <div>
                        <h2
                            className="modal__title"
                            id="complete-quarter-modal-title"
                        >
                            Завершение четверти
                        </h2>

                        <p className="modal__description">
                            Подтвердите завершение учебного периода для определения лучшего класса
                        </p>
                    </div>

                    <button
                        className="modal__close"
                        type="button"
                        onClick={() => {
                            onClose();
                        }}
                    >
                        ×
                    </button>
                </div>

                <div className="modal__body">

                    {!completed && (
                        <div className="alert alert--warning">
                            Вы уверены, что хотите завершить учебный период?
                        </div>
                    )}

                    {isError && (
                        <div className="alert alert--danger">
                            {error?.message}
                        </div>
                    )}

                    {completed && data && (
                        <div className="quarter-results">

                            <div className="quarter-results__grid">

                                {renderClassPlace(
                                    "1 Место",
                                    data["1st"]
                                )}

                                {renderClassPlace(
                                    "2 Место",
                                    data["2nd"]
                                )}

                                {renderClassPlace(
                                    "3 Место",
                                    data["3rd"]
                                )}

                            </div>
                        </div>
                    )}

                </div>

                <div className="modal__footer">

                    <button
                        className="btn btn--secondary"
                        type="button"
                        onClick={() => {
                            onClose();
                        }}
                        disabled={isPending}
                    >
                        {completed ? "Закрыть" : "Отмена"}
                    </button>

                    {!completed && (
                        <button
                            className="btn btn--primary"
                            type="button"
                            onClick={handleCompleteQuarter}
                            disabled={isPending}
                        >
                            {isPending
                                ? "Завершение..."
                                : "Подтвердить"}
                        </button>
                    )}

                </div>

            </section>
        </div>
    );
}