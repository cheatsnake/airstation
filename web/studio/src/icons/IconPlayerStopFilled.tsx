import { FC } from "react";
import { IconProps } from "./types";

export const IconPlayerStopFilled: FC<IconProps> = ({
  width = 24,
  height = 24,
  stroke = 1.5,
  fill,
  style,
  className,
}) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width={width}
      height={height}
      viewBox="0 0 24 24"
      fill={fill}
      strokeWidth={stroke}
      className={className}
      style={style}
    >
      <path stroke="none" d="M0 0h24v24H0z" fill="none" />
      <path d="M17 4h-10a3 3 0 0 0 -3 3v10a3 3 0 0 0 3 3h10a3 3 0 0 0 3 -3v-10a3 3 0 0 0 -3 -3z" />
    </svg>
  );
};
