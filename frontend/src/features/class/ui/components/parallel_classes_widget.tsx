import {useNavigate} from "react-router-dom";
import {ClassCard} from "../../../../shared/ui/cards/class_card.tsx";
import type {ClassType} from "../../../../shared/entities/class/types/class_types.ts";
import {useParallelClasses} from "../../hooks/use_parallel_classes.ts";

export function ParallelClassesWidget({id}: {id: number}) {
    const navigate = useNavigate();
    const {data, isLoading, error} = useParallelClasses(id);
    const classes = data as ClassType[];

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
                <div className="alert alert--danger mb-4">{error.message}</div>
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
                        Не удалось найти классы по параллели
                    </p>
                </div>
            )}
        </>
    );
} 