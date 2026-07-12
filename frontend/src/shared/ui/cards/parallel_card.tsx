import type { ParallelType } from "../../entities/class/types/class_types.ts";
import { useParallelClasses } from "../../../features/class/hooks/use_parallel_classes.ts";

interface Props {
    item: ParallelType;
    onClick?: () => void;
}

export function ParallelCard({ item, onClick }: Props) {
    const classesCount = item.ClassesIds?.length ?? 0;

    const { data: classes } = useParallelClasses(item.Id);

    const maxVal = Math.max(
        ...(classes?.map(eachClass => (eachClass?.ClassTotalRating ?? 0) + (eachClass?.UserTotalRating ?? 0)) ?? [0]));

    const bestClass = classes?.find(
        eachClass => (eachClass?.ClassTotalRating ?? 0) + (eachClass?.UserTotalRating ?? 0) === maxVal
    );

    return (
        <button
            className="class-flat-card"
            type="button"
            onClick={onClick}
        >
            <div className="class-flat-card__main">
                <div className="class-flat-card__info">
                    <h2 className="class-flat-card__title">
                        {item.Name}
                    </h2>

                    <p className="class-flat-card__subtitle">
                        Количество классов: {classesCount}
                    </p>

                </div>
            </div>

            <div className="class-flat-card__metric">
                <span className="class-flat-card__metric-label">Кол-во классов</span>
                <span className="class-flat-card__metric-value">{classesCount}</span>
            </div>
            
            <div className="class-flat-card__metric">
                <span className="class-flat-card__metric-label">Лучший класс</span>
                <span className="class-flat-card__metric-value">{bestClass?.Name ?? "Нет"}</span>
            </div>
        </button>
    );
}