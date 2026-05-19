import {useNavigate} from "react-router-dom";
import {useEffect} from "react";
import {useClassStore} from "../../store/class_store.ts";
import {ClassCard} from "../../../../shared/ui/cards/class_card.tsx";

export function ClassesWidget() {
    const navigate = useNavigate();
    const classes = useClassStore((state) => state.classes);
    const getClasses = useClassStore((state) => state.getClasses);
    const status = useClassStore((state) => state.status);
    const error = useClassStore((state) => state.error);

    const isLoading = status === "loading";

    useEffect(() => {
        void getClasses();
    }, [getClasses])

    return (
        <>
            {isLoading && (
                <div className="grid grid--3">
                    <div className="skeleton" style={{ height: 160 }} />
                    <div className="skeleton" style={{ height: 160 }} />
                    <div className="skeleton" style={{ height: 160 }} />
                </div>
            )}

            {error && !isLoading && (
                <div className="alert alert--danger mb-4">{error}</div>
            )}

            {classes && !isLoading && !error && classes.length > 0 ? (
                <div className="class-list">
                    {classes.map((item) => (
                        <ClassCard key={item.Id} item={item} onClick={() => {
                            navigate(`/class/${item.Id}?name=${item.Name}`);
                        }}/>
                    ))}
                </div>
            ) : (
                !isLoading && !error && (classes?.length === 0 || !classes) && <div className="empty-state">
                    <h2 className="empty-state__title">Классы не найдены</h2>
                    <p className="empty-state__text">
                        В системе пока нет доступных классов
                    </p>
                </div>
            )}
        </>
    );
} 